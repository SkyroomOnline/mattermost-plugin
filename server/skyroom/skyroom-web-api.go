package skyroom

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	jsonitor "github.com/json-iterator/go"
)

type WebAPI interface {
	GetServices() ([]*Service, error)
	CreateRoomIfNotExists(name, title string) (*RoomInfo, error)
	CreateLoginURL(roomId int, userId, nickname string, ttl int) (string, error)
}

type APIError struct {
	code    int
	message string
}

func (err *APIError) Error() string {
	return fmt.Sprintf("Code %d: %s", err.code, err.message)
}

var (
	ErrorInvalidAPIKey  = &APIError{code: 11, message: "invalid-api-key"}
	ErrorInvalidRequest = &APIError{code: 12, message: "invalid-request"}
	ErrorFailed         = &APIError{code: 14, message: "failed"}
	ErrorNotFound       = &APIError{code: 15, message: "not-found"}
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

func makeErrorFromCode(code int) error {
	if code == 10 || code == 11 {
		return ErrorInvalidAPIKey
	} else if code == 12 || code == 13 {
		return ErrorInvalidRequest
	} else if code == 14 {
		return ErrorFailed
	} else if code == 15 {
		return ErrorNotFound
	} else {
		return ErrorFailed
	}
}

func MakeWebAPI(url, apiKey string) WebAPI {
	return &skyroomWebAPI{URL: url, APIKey: apiKey}
}

func (skyroom *skyroomWebAPI) sendRequest(action string, params map[string]interface{}) ([]byte, error) {
	url1 := fmt.Sprintf("%s/skyroom/api/%s", skyroom.URL, skyroom.APIKey)
	data := make(map[string]interface{})
	data["action"] = action
	if len(params) > 0 {
		data["params"] = params
	}

	body, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	resp, postErr := http.Post(url1, "application/json", bytes.NewReader(body))
	if postErr != nil {
		return nil, postErr
	}
	respBody, readErr := ioutil.ReadAll(resp.Body)
	if readErr != nil {
		return nil, readErr
	}
	return respBody, nil
}

func (skyroom *skyroomWebAPI) GetServices() ([]*Service, error) {
	var list []*Service
	data, err := skyroom.sendRequest("getServices", nil)
	if err != nil {
		return list, err
	}
	jsonErr := json.Unmarshal(data, &list)
	if jsonErr != nil {
		return list, jsonErr
	}
	return list, nil
}

func (skyroom *skyroomWebAPI) getRoomByName(name string) (*RoomInfo, error) {
	var room RoomInfo
	data, err := skyroom.sendRequest("getRoom", map[string]interface{}{"name": name})
	if err != nil {
		return nil, err
	}
	ok := jsonitor.Get(data, "ok").ToBool()
	if !ok {
		code := jsonitor.Get(data, "error_code").ToInt()
		return nil, makeErrorFromCode(code)
		// return nil, errors.New(jsonitor.Get(data, "error_message").ToString())
	}
	jsonResult := jsonitor.Get(data, "result").ToString()
	jsonErr := jsonitor.UnmarshalFromString(jsonResult, &room)
	if jsonErr != nil {
		return nil, jsonErr
	}
	return &room, nil
}

func (skyroom *skyroomWebAPI) createRoom(name, title string) error {
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
		return makeErrorFromCode(code)
		// return errors.New(jsonitor.Get(data, "error_message").ToString())
	}
	return nil
}

func (skyroom *skyroomWebAPI) CreateRoomIfNotExists(name, title string) (*RoomInfo, error) {
	room, err := skyroom.getRoomByName(name)
	if err.Error() == ErrorNotFound.Error() {
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

func (skyroom *skyroomWebAPI) CreateLoginURL(roomId int, userId, nickname string, ttl int) (string, error) {
	params := map[string]interface{}{
		"room_id":    roomId,
		"user_id":    userId,
		"nickname":   nickname,
		"access":     3,
		"concurrent": 1,
		"language":   "en",
		"ttl":        ttl,
	}
	data, err := skyroom.sendRequest("createLoginUrl", params)
	if err != nil {
		return "", err
	}
	ok := jsonitor.Get(data, "ok").ToBool()
	if !ok {
		code := jsonitor.Get(data, "error_code").ToInt()
		return "", makeErrorFromCode(code)
		// return nil, errors.New(jsonitor.Get(data, "error_message").ToString())
	}
	link := jsonitor.Get(data, "result").ToString()
	return link, nil
}
