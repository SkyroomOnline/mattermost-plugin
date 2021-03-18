package main

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"github.com/mattermost/mattermost-plugin-api/i18n"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
	"github.com/pkg/errors"
	"github.com/skyroomonline/mattermost-plugin-skyroom/server/skyroom"
)

const skyroomNameSchemeAsk = "ask"
const skyroomNameSchemeWords = "words"
const skyroomNameSchemeUUID = "uuid"
const skyroomNameSchemeMattermost = "mattermost"
const configChangeEvent = "config_update"

type UserConfig struct {
	Embedded     bool   `json:"embedded"`
	NamingScheme string `json:"naming_scheme"`
}

type Plugin struct {
	plugin.MattermostPlugin

	// configurationLock synchronizes access to the configuration.
	configurationLock sync.RWMutex

	// configuration is the active plugin configuration. Consult getConfiguration and
	// setConfiguration for usage.
	configuration *configuration

	skyroomAPI skyroom.WebAPI

	b *i18n.Bundle

	botID string
}

func (p *Plugin) OnActivate() error {
	config := p.getConfiguration()
	if err := config.IsValid(); err != nil {
		return err
	}

	if p.skyroomAPI == nil {
		p.skyroomAPI = skyroom.MakeWebAPI(config.GetSkyroomURL(), config.SkyroomAPIKey)
	}

	command, err := p.createSkyroomCommand()
	if err != nil {
		return err
	}

	if err = p.API.RegisterCommand(command); err != nil {
		return err
	}

	i18nBundle, err := i18n.InitBundle(p.API, filepath.Join("assets", "i18n"))
	if err != nil {
		return err
	}
	p.b = i18nBundle

	skyroomBot := &model.Bot{
		Username:    "skyroom",
		DisplayName: "skyroom",
		Description: "A bot account created by the skyroom plugin",
	}
	options := []plugin.EnsureBotOption{
		plugin.ProfileImagePath("assets/icon.png"),
	}

	botID, ensureBotError := p.Helpers.EnsureBot(skyroomBot, options...)
	if ensureBotError != nil {
		return errors.Wrap(ensureBotError, "failed to ensure skyroom bot user.")
	}

	p.botID = botID

	return nil
}

type User struct {
	Avatar string `json:"avatar"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	ID     string `json:"id"`
}

type InvitationContext struct {
	RoomId    int    `json:"roomId"`
	ChannelId string `json:"channelId"`
	Expires   int64  `json:"expires"`
}

func (p *Plugin) startMeeting(user *model.User, channel *model.Channel, rootID string) (string, error) {
	// l := p.b.GetServerLocalizer()
	configs := p.getConfiguration()
	roomId := fmt.Sprintf("mattermost-%s", channel.Id)

	skyRoom, roomErr := p.skyroomAPI.CreateRoomIfNotExists(roomId, channel.DisplayName)
	if roomErr != nil {
		return "", roomErr
	}

	meetingLinkValidUntil := time.Now().Add(30 * time.Minute)

	params := InvitationContext{
		RoomId:    skyRoom.Id,
		ChannelId: channel.Id,
		Expires:   meetingLinkValidUntil.Unix(),
	}
	jsonParams, _ := json.Marshal(params)
	encrypted, _ := Encrypt([]byte(configs.GetEncryptionKey()), string(jsonParams))

	invitationURL := fmt.Sprintf("/plugins/online.skyroom.skyroom/api/v1/join?p=%s", encrypted) //plugins/online.skyroom.skyroom/api/v1/join

	post := &model.Post{
		UserId:    user.Id,
		ChannelId: channel.Id,
		Type:      "custom_skyroom",
		Props: map[string]interface{}{
			"meeting_id":            skyRoom.Id,
			"join_parameter":        encrypted,
			"valid_until":           meetingLinkValidUntil.Unix(),
			"meeting_join_url":      invitationURL,
			"meeting_topic":         channel.DisplayName,
			"default_meeting_topic": channel.DisplayName,
		},
		RootId: rootID,
	}

	if _, err := p.API.CreatePost(post); err != nil {
		return "", err
	}

	return invitationURL, nil
}

func (p *Plugin) joinMeeting(user *model.User, parameters string) (string, error) {
	if user == nil {
		return "", errors.New("no-user-found")
	}
	configs := p.getConfiguration()
	decrypted, _ := Decrypt([]byte(configs.GetEncryptionKey()), parameters)
	var invitationctx InvitationContext
	jsonErr := json.Unmarshal([]byte(decrypted), &invitationctx)
	if jsonErr != nil {
		return "", jsonErr
	}
	// channel, chErr := p.API.GetChannel(invitationctx.ChannelId)
	// if chErr != nil {
	// 	return "", chErr
	// }
	meetingLinkValidUntil := int64(invitationctx.Expires)
	if time.Now().Unix() > meetingLinkValidUntil {
		return "", errors.New("link-has-expired")
	}
	skyRoomId := invitationctx.RoomId
	nickName := user.GetFullName()
	if len(nickName) < 1 {
		nickName = user.Nickname
	}
	if len(nickName) < 1 {
		nickName = user.Username
	}

	skyroomLink, linkErr := p.skyroomAPI.CreateLoginURL(skyRoomId, user.Id, nickName, configs.SkyroomLinkValidTime*60)
	if linkErr != nil {
		return "", linkErr
	}
	return skyroomLink, nil

	// post := &model.Post{
	// 	UserId:    p.botID,
	// 	ChannelId: channel.Id,
	// 	Message:   errorText,
	// 	RootId:    "",
	// }
	// _ = p.API.SendEphemeralPost(user.Id, post)

}
