package main

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/mattermost/mattermost-plugin-api/i18n"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin/plugintest"
	"github.com/skyroomonline/mattermost-plugin-skyroom/server/skyroom"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestStartMeeting(t *testing.T) {
	p := Plugin{
		configuration: &configuration{
			SkyroomURL:           "https://dev.skyroom.online",
			SkyroomAPIKey:        "apikey-2-9927693-c8db213558ac0f0d3ca2bdc9c0ed66b2",
			SkyroomLinkValidTime: 30,
			EncryptionKey:        "sj9xLxNBRiE1RN6d",
		},
	}

	if p.skyroomAPI == nil {
		p.skyroomAPI = skyroom.MakeWebAPI(p.configuration.SkyroomURL, p.configuration.SkyroomAPIKey)
	}
	apiMock := plugintest.API{}
	defer apiMock.AssertExpectations(t)
	apiMock.On("GetBundlePath").Return("..", nil)
	// config := model.Config{}
	// config.SetDefaults()
	// apiMock.On("GetConfig").Return(&config, nil)

	p.SetAPI(&apiMock)

	i18nBundle, err := i18n.InitBundle(p.API, filepath.Join("assets", "i18n"))
	require.Nil(t, err)
	p.b = i18nBundle

	testUser := model.User{Id: "test-id", Username: "test-username", FirstName: "test-first-name", LastName: "test-last-name", Nickname: "test-nickname"}
	testChannel := model.Channel{Id: "test-id", Type: model.CHANNEL_DIRECT, Name: "test-name", DisplayName: "test-display-name"}
	var joinParameters string
	t.Run("start meeting for test user in test channel", func(t *testing.T) {
		apiMock.On("CreatePost", mock.MatchedBy(func(post *model.Post) bool {
			joinParameters = post.Props["join_parameter"].(string)
			return len(joinParameters) > 10
		})).Return(&model.Post{}, nil)
		invitationURL, err := p.startMeeting(&testUser, &testChannel, "")
		require.Nil(t, err)
		require.Regexp(t, `^\/plugins\/online\.skyroom\.skyroom\/api\/v1\/join\?p=.*`, invitationURL)
	})

	t.Run("join meeting using generated login link of skyroom", func(t *testing.T) {

		skyroomLoginLink, err := p.joinMeeting(&testUser, joinParameters)
		require.Nil(t, err)
		require.Regexp(t, fmt.Sprintf("^%s/ch/mattermost-%s", p.configuration.SkyroomURL, testChannel.Id), skyroomLoginLink)
	})
}
