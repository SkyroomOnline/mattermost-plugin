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
}

type WebAPI interface {
	GetServices() ([]*Service, APIError)
	CreateRoomIfNotExists(name, title string) (*RoomInfo, APIError)
	CreateLoginURL(roomId int, userId, nickname string, ttl int) (string, APIError)
}

type apiError struct {
	req           map[string]interface{}
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
	jsonErr["internalError"] = err.internalError
	body, _ := json.Marshal(jsonErr)
	return string(body)
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

func makeAPIErrorFromError(internal error, action string, params map[string]interface{}) APIError {
	apiErr := apiError{code: 14, message: "failed"}
	apiErr.req = make(map[string]interface{})
	apiErr.internalError = internal.Error()
	return &apiErr
}

func makeAPIErrorFromCode(code int, action string, params map[string]interface{}) APIError {
	data := make(map[string]interface{})
	data["action"] = action
	if len(params) > 0 {
		data["params"] = params
	}
	if code == 10 || code == 11 {
		err := apiError{code: 11, message: "invalid-api-key"}
		err.req = data
		return &err
	} else if code == 12 || code == 13 {
		err := apiError{code: 12, message: "invalid-request"}
		err.req = data
		return &err
	} else if code == 14 {
		err := apiError{code: 14, message: "failed"}
		err.req = data
		return &err
	} else if code == 15 {
		err := apiError{code: 15, message: "not-found"}
		err.req = data
		return &err
	} else {
		err := apiError{code: 14, message: "failed"}
		err.req = data
		return &err
	}
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
		return nil, makeAPIErrorFromError(err, action, params)
	}

	resp, postErr := http.Post(url1, "application/json", bytes.NewReader(body))
	if postErr != nil {
		return nil, makeAPIErrorFromError(postErr, action, params)
	}
	respBody, readErr := ioutil.ReadAll(resp.Body)
	if readErr != nil {
		return nil, makeAPIErrorFromError(readErr, action, params)
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
		return list, makeAPIErrorFromError(jsonErr, "getServices", nil)
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
		code := jsonitor.Get(data, "error_code").ToInt()
		return nil, makeAPIErrorFromCode(code, "getRoom", params)
		// return nil, errors.New(jsonitor.Get(data, "error_message").ToString())
	}
	jsonResult := jsonitor.Get(data, "result").ToString()
	jsonErr := jsonitor.UnmarshalFromString(jsonResult, &room)
	if jsonErr != nil {
		return nil, makeAPIErrorFromError(jsonErr, "getRoom", nil)
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
		code := jsonitor.Get(data, "error_code").ToInt()
		return makeAPIErrorFromCode(code, "createRoom", params)
		// return errors.New(jsonitor.Get(data, "error_message").ToString())
	}
	return nil
}

func (skyroom *skyroomWebAPI) CreateRoomIfNotExists(name, title string) (*RoomInfo, APIError) {
	room, err := skyroom.getRoomByName(name)
	if err.Code() == ErrorNotFound.Code() {
		_ = skyroom.createRoom(name, title)
		room, err = skyroom.getRoomByName(name)
		if err != nil {
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
		code := jsonitor.Get(data, "error_code").ToInt()
		return "", makeAPIErrorFromCode(code, "createLoginUrl", params)
		// return nil, errors.New(jsonitor.Get(data, "error_message").ToString())
	}
	link := jsonitor.Get(data, "result").ToString()
	return link, nil
}
