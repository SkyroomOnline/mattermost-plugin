package skyroom

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	jsonitor "github.com/json-iterator/go"
)

type APIError interface {
	error
	Code() int
	SetInternalError(err error)
}

type WebAPI interface {
	GetServices() ([]*Service, APIError)
	CreateRoomIfNotExists(name, title string) (*RoomInfo, APIError)
	CreateLoginURL(roomId int, userId, nickname string, ttl int) (string, APIError)
}

type apiError struct {
	req           map[string]interface{}
	resp          string
	code          int
	message       string
	internalError string
}

func (err *apiError) Code() int {
	return err.code
}

func (err *apiError) Error() string {
	jsonErr := make(map[string]interface{})
	jsonErr["request"] = err.req
	jsonErr["error"] = fmt.Sprintf("Code %d: %s", err.code, err.message)
	jsonErr["resp"] = err.resp
	jsonErr["internalError"] = err.internalError
	body, _ := json.Marshal(jsonErr)
	return string(body)
}

func (err *apiError) SetInternalError(internalErr error) {
	err.internalError = internalErr.Error()
}

var (
	ErrorInvalidAPIKey  = apiError{code: 11, message: "invalid-api-key"}
	ErrorInvalidRequest = apiError{code: 12, message: "invalid-request"}
	ErrorFailed         = apiError{code: 14, message: "failed"}
	ErrorNotFound       = apiError{code: 15, message: "not-found"}
)

type skyroomWebAPI struct {
	URL    string
	APIKey string
}

type skyroomError struct {
	OK           bool   `json:"ok"`
	ErrorCode    int    `json:"error_code"`
	ErrorMessage string `json:"error_message"`
}

type Service struct {
	Id         int    `json:"id"`
	Title      string `json:"title"`
	Status     int    `json:"status"`
	UserLimit  int    `json:"user_limit"`
	VideoLimit int    `json:"video_limit"`
	//TimeLimit `json:"time_limit"`
	TimeUsage  int   `json:"time_usage"`
	StartTime  int64 `json:"start_time"`
	StopTime   int64 `json:"stop_time"`
	CreateTime int64 `json:"create_time"`
	UpdateTime int64 `json:"update_time"`
}

type RoomInfo struct {
	Id              int    `json:"id"`
	ServiceId       int    `json:"service_id"`
	Name            string `json:"name"`
	Title           string `json:"title"`
	Description     string `json:"description"`
	Status          int    `json:"status"`
	GuestLogin      bool   `json:"guest_login"`
	GuestLimit      int    `json:"guest_limit"`
	OpLoginFirst    bool   `json:"op_login_first"`
	MaxUsers        int    `json:"max_users"`
	SessionDuration int    `json:"session_duration"`
	TimeLimit       int    `json:"time_limit"`
	TimeUsage       int    `json:"time_usage"`
	TimeTotal       int    `json:"time_total"`
	CreateTime      int64  `json:"create_time"`
	UpdateTime      int64  `json:"update_time"`
}

func makeAPIErrorFromError(internal error, action string, params map[string]interface{}, resp string) APIError {
	apiErr := apiError{code: 14, message: "failed"}
	apiErr.resp = resp
	apiErr.req = make(map[string]interface{})
	apiErr.internalError = internal.Error()
	return &apiErr
}

func makeAPIErrorFromCode(code int, action string, params map[string]interface{}, resp string) APIError {
	data := make(map[string]interface{})
	data["action"] = action
	if len(params) > 0 {
		data["params"] = params
	}
	err := apiError{code: code, message: "failed"}
	if code == 10 || code == 11 {
		err = apiError{code: 11, message: "invalid-api-key"}
	} else if code == 12 || code == 13 {
		err = apiError{code: 12, message: "invalid-request"}
	} else if code == 14 {
		err = apiError{code: 14, message: "failed"}
	} else if code == 15 {
		err = apiError{code: 15, message: "not-found"}
	}
	err.req = data
	err.resp = resp
	return &err
}

func MakeWebAPI(url, apiKey string) WebAPI {
	return &skyroomWebAPI{URL: url, APIKey: apiKey}
}

func (skyroom *skyroomWebAPI) sendRequest(action string, params map[string]interface{}) ([]byte, APIError) {
	url1 := fmt.Sprintf("%s/skyroom/api/%s", skyroom.URL, skyroom.APIKey)
	data := make(map[string]interface{})
	data["action"] = action
	if len(params) > 0 {
		data["params"] = params
	}

	body, err := json.Marshal(data)
	if err != nil {
		return nil, makeAPIErrorFromError(err, action, params, "")
	}

	resp, postErr := http.Post(url1, "application/json", bytes.NewReader(body))
	if postErr != nil {
		return nil, makeAPIErrorFromError(postErr, action, params, "")
	}
	respBody, readErr := ioutil.ReadAll(resp.Body)
	if readErr != nil {
		return nil, makeAPIErrorFromError(readErr, action, params, "")
	}
	return respBody, nil
}

func (skyroom *skyroomWebAPI) GetServices() ([]*Service, APIError) {
	var list []*Service
	data, err := skyroom.sendRequest("getServices", nil)
	if err != nil {
		return list, err
	}
	jsonErr := json.Unmarshal(data, &list)
	if jsonErr != nil {
		return list, makeAPIErrorFromError(jsonErr, "getServices", nil, string(data))
	}
	return list, nil
}

func (skyroom *skyroomWebAPI) getRoomByName(name string) (*RoomInfo, APIError) {
	var room RoomInfo
	params := map[string]interface{}{"name": name}
	data, err := skyroom.sendRequest("getRoom", params)
	if err != nil {
		return nil, err
	}
	ok := jsonitor.Get(data, "ok").ToBool()
	if !ok {
		var skyroomErr skyroomError
		if jsonskyErr := jsonitor.Unmarshal(data, &skyroomErr); jsonskyErr != nil {
			return nil, makeAPIErrorFromError(jsonskyErr, "getRoom", params, string(data))
		}
		// code := int(jsonitor.Get(data, "error_code").ToFloat64())
		return nil, makeAPIErrorFromCode(skyroomErr.ErrorCode, "getRoom", params, string(data))
		// return nil, errors.New(jsonitor.Get(data, "error_message").ToString())
	}
	jsonResult := jsonitor.Get(data, "result").ToString()
	jsonErr := jsonitor.UnmarshalFromString(jsonResult, &room)
	if jsonErr != nil {
		return nil, makeAPIErrorFromError(jsonErr, "getRoom", params, string(data))
	}
	return &room, nil
}

func (skyroom *skyroomWebAPI) createRoom(name, title string) APIError {
	params := map[string]interface{}{
		"name":           name,
		"title":          title,
		"guest_login":    false,
		"op_login_first": true,
		"max_users":      50,
	}
	data, err := skyroom.sendRequest("createRoom", params)
	if err != nil {
		return err
	}
	ok := jsonitor.Get(data, "ok").ToBool()
	if !ok {
		var skyroomErr skyroomError
		if jsonskyErr := jsonitor.Unmarshal(data, &skyroomErr); jsonskyErr != nil {
			return makeAPIErrorFromError(jsonskyErr, "createRoom", params, string(data))
		}
		//code := jsonitor.Get(data, "error_code").ToInt()
		return makeAPIErrorFromCode(skyroomErr.ErrorCode, "createRoom", params, string(data))
		// return errors.New(jsonitor.Get(data, "error_message").ToString())
	}
	return nil
}

func (skyroom *skyroomWebAPI) CreateRoomIfNotExists(name, title string) (*RoomInfo, APIError) {
	room, err := skyroom.getRoomByName(name)
	if err != nil && err.Code() == ErrorNotFound.Code() {
		creationError := skyroom.createRoom(name, title)
		room, err = skyroom.getRoomByName(name)
		if err != nil {
			err.SetInternalError(creationError)
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}
	return room, nil

}

func (skyroom *skyroomWebAPI) CreateLoginURL(roomId int, userId, nickname string, ttl int) (string, APIError) {
	params := map[string]interface{}{
		"room_id":    roomId,
		"user_id":    userId,
		"nickname":   nickname,
		"access":     3,
		"concurrent": 1,
		"language":   "fa",
		"ttl":        ttl,
	}
	data, err := skyroom.sendRequest("createLoginUrl", params)
	if err != nil {
		return "", err
	}
	ok := jsonitor.Get(data, "ok").ToBool()
	if !ok {
		var skyroomErr skyroomError
		if jsonskyErr := jsonitor.Unmarshal(data, &skyroomErr); jsonskyErr != nil {
			return "", makeAPIErrorFromError(jsonskyErr, "createLoginUrl", params, string(data))
		}
		// code := jsonitor.Get(data, "error_code").ToInt()
		return "", makeAPIErrorFromCode(skyroomErr.ErrorCode, "createLoginUrl", params, string(data))
		// return nil, errors.New(jsonitor.Get(data, "error_message").ToString())
	}
	link := jsonitor.Get(data, "result").ToString()
	return link, nil
}
