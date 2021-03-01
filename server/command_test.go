package main

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mattermost/mattermost-plugin-api/i18n"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
	"github.com/mattermost/mattermost-server/v5/plugin/plugintest"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCommandHelp(t *testing.T) {
	p := Plugin{
		configuration: &configuration{
			SkyroomURL: "http://test",
		},
		botID: "test-bot-id",
	}
	apiMock := plugintest.API{}
	defer apiMock.AssertExpectations(t)
	apiMock.On("GetUser", "test-user").Return(&model.User{Id: "test-user", Locale: "en"}, nil)
	apiMock.On("GetBundlePath").Return("..", nil)

	p.SetAPI(&apiMock)

	i18nBundle, err := i18n.InitBundle(p.API, filepath.Join("assets", "i18n"))
	require.Nil(t, err)
	p.b = i18nBundle

	helpText := strings.ReplaceAll(`###### Mattermost Skyroom Plugin - Slash Command help
* |/skyroom| - Create a new meeting and invite others
* |/skyroom start| - Create a new meeting and invite others
* |/skyroom help| - Show this help text`, "|", "`")

	apiMock.On("SendEphemeralPost", "test-user", &model.Post{
		UserId:    "test-bot-id",
		ChannelId: "test-channel",
		Message:   helpText,
	}).Return(nil)
	response, err := p.ExecuteCommand(&plugin.Context{}, &model.CommandArgs{UserId: "test-user", ChannelId: "test-channel", Command: "/skyroom help"})
	require.Equal(t, &model.CommandResponse{}, response)
	require.Nil(t, err)
}

func TestCommandStartMeeting(t *testing.T) {
	p := Plugin{
		configuration: &configuration{
			SkyroomURL: "http://test",
		},
	}

	t.Run("meeting without topic and ask configuration", func(t *testing.T) {
		apiMock := plugintest.API{}
		defer apiMock.AssertExpectations(t)
		p.SetAPI(&apiMock)

		apiMock.On("GetBundlePath").Return("..", nil)
		config := model.Config{}
		config.SetDefaults()
		apiMock.On("GetConfig").Return(&config, nil)

		i18nBundle, err := i18n.InitBundle(p.API, filepath.Join("assets", "i18n"))
		require.Nil(t, err)
		p.b = i18nBundle

		apiMock.On("SendEphemeralPost", "test-user", mock.MatchedBy(func(post *model.Post) bool {
			return post.Props["attachments"].([]*model.SlackAttachment)[0].Text == "Select type of meeting you want to start"
		})).Return(nil)
		apiMock.On("GetUser", "test-user").Return(&model.User{Id: "test-user"}, nil)
		apiMock.On("GetChannel", "test-channel").Return(&model.Channel{Id: "test-channel"}, nil)
		apiMock.On("GetUser", "test-user").Return(&model.User{Id: "test-user"}, nil)
		b, _ := json.Marshal(UserConfig{Embedded: false, NamingScheme: "ask"})
		apiMock.On("KVGet", "config_test-user", mock.Anything).Return(b, nil)

		response, err := p.ExecuteCommand(&plugin.Context{}, &model.CommandArgs{UserId: "test-user", ChannelId: "test-channel", Command: "/skyroom"})
		require.Equal(t, &model.CommandResponse{}, response)
		require.Nil(t, err)
	})

	t.Run("meeting without topic and no ask configuration", func(t *testing.T) {
		apiMock := plugintest.API{}
		defer apiMock.AssertExpectations(t)
		p.SetAPI(&apiMock)

		apiMock.On("GetBundlePath").Return("..", nil)
		config := model.Config{}
		config.SetDefaults()
		apiMock.On("GetConfig").Return(&config, nil)

		i18nBundle, err := i18n.InitBundle(p.API, filepath.Join("assets", "i18n"))
		require.Nil(t, err)
		p.b = i18nBundle

		apiMock.On("CreatePost", mock.MatchedBy(func(post *model.Post) bool {
			return strings.HasPrefix(post.Props["meeting_link"].(string), "http://test/")
		})).Return(&model.Post{}, nil)
		apiMock.On("GetUser", "test-user").Return(&model.User{Id: "test-user"}, nil)
		apiMock.On("GetChannel", "test-channel").Return(&model.Channel{Id: "test-channel"}, nil)
		apiMock.On("GetUser", "test-user").Return(&model.User{Id: "test-user"}, nil)
		apiMock.On("KVGet", "config_test-user", mock.Anything).Return(nil, nil)

		response, err := p.ExecuteCommand(&plugin.Context{}, &model.CommandArgs{UserId: "test-user", ChannelId: "test-channel", Command: "/skyroom"})
		require.Equal(t, &model.CommandResponse{}, response)
		require.Nil(t, err)
	})

}
