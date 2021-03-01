package main

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mattermost/mattermost-plugin-api/i18n"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin/plugintest"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestStartMeeting(t *testing.T) {
	p := Plugin{
		configuration: &configuration{
			SkyroomURL: "http://test",
		},
	}
	apiMock := plugintest.API{}
	defer apiMock.AssertExpectations(t)
	apiMock.On("GetBundlePath").Return("..", nil)
	config := model.Config{}
	config.SetDefaults()
	apiMock.On("GetConfig").Return(&config, nil)

	p.SetAPI(&apiMock)

	i18nBundle, err := i18n.InitBundle(p.API, filepath.Join("assets", "i18n"))
	require.Nil(t, err)
	p.b = i18nBundle

	testUser := model.User{Id: "test-id", Username: "test-username", FirstName: "test-first-name", LastName: "test-last-name", Nickname: "test-nickname"}
	testChannel := model.Channel{Id: "test-id", Type: model.CHANNEL_DIRECT, Name: "test-name", DisplayName: "test-display-name"}

	t.Run("start meeting without topic or id", func(t *testing.T) {
		apiMock.On("CreatePost", mock.MatchedBy(func(post *model.Post) bool {
			return strings.HasPrefix(post.Props["meeting_link"].(string), "http://test/")
		})).Return(&model.Post{}, nil)
		b, _ := json.Marshal(UserConfig{Embedded: false, NamingScheme: "mattermost"})
		apiMock.On("KVGet", "config_test-id", mock.Anything).Return(b, nil)

		meetingID, err := p.startMeeting(&testUser, &testChannel, "")
		require.Nil(t, err)
		require.Regexp(t, "^test-username-", meetingID)
	})

	t.Run("start meeting with topic and without id", func(t *testing.T) {
		apiMock.On("CreatePost", mock.MatchedBy(func(post *model.Post) bool {
			return strings.HasPrefix(post.Props["meeting_link"].(string), "http://test/")
		})).Return(&model.Post{}, nil)
		b, _ := json.Marshal(UserConfig{Embedded: false, NamingScheme: "mattermost"})
		apiMock.On("KVGet", "config_test-id", mock.Anything).Return(b, nil)

		meetingID, err := p.startMeeting(&testUser, &testChannel, "")
		require.Nil(t, err)
		require.Regexp(t, "^Test-topic-", meetingID)
	})

	t.Run("start meeting without topic and with id", func(t *testing.T) {
		apiMock.On("CreatePost", mock.MatchedBy(func(post *model.Post) bool {
			return strings.HasPrefix(post.Props["meeting_link"].(string), "http://test/")
		})).Return(&model.Post{}, nil)
		b, _ := json.Marshal(UserConfig{Embedded: false, NamingScheme: "mattermost"})
		apiMock.On("KVGet", "config_test-id", mock.Anything).Return(b, nil)

		meetingID, err := p.startMeeting(&testUser, &testChannel, "")
		require.Nil(t, err)
		require.Regexp(t, "^test-username-", meetingID)
	})

	t.Run("start meeting with topic and id", func(t *testing.T) {
		testUser := model.User{Id: "test-id", Username: "test-username", FirstName: "test-first-name", LastName: "test-last-name", Nickname: "test-nickname"}
		testChannel := model.Channel{Id: "test-id", Type: model.CHANNEL_OPEN, TeamId: "test-team-id", Name: "test-name", DisplayName: "test-display-name"}

		apiMock.On("CreatePost", mock.MatchedBy(func(post *model.Post) bool {
			return strings.HasPrefix(post.Props["meeting_link"].(string), "http://test/")
		})).Return(&model.Post{}, nil)
		b, _ := json.Marshal(UserConfig{Embedded: false, NamingScheme: "mattermost"})
		apiMock.On("KVGet", "config_test-id", mock.Anything).Return(b, nil)

		meetingID, err := p.startMeeting(&testUser, &testChannel, "")
		require.Nil(t, err)
		require.Equal(t, "test-id", meetingID)
	})
}
