package main

import (
	"fmt"
	"strings"

	"github.com/mattermost/mattermost-plugin-api/experimental/command"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/pkg/errors"
)

const skyroomCommand = "skyroom"

const skyroomSettingsCommand = "settings"
const skyroomStartCommand = "start"

func (p *Plugin) createSkyroomCommand() (*model.Command, error) {
	iconData, err := command.GetIconData(p.API, "assets/icon.svg")
	if err != nil {
		return nil, errors.Wrap(err, "failed to get icon data")
	}
	return &model.Command{
		Trigger:              skyroomCommand,
		AutoComplete:         true,
		AutoCompleteDesc:     "Start a skyroom meeting for current channel and notify others. Other available commands: start, help, settings",
		AutoCompleteHint:     "[command]",
		AutocompleteData:     getAutocompleteData(),
		AutocompleteIconData: iconData,
	}, nil
}

func getAutocompleteData() *model.AutocompleteData {
	skyroom := model.NewAutocompleteData(skyroomCommand, "[command]", "Start a skyroom meeting for current channel and notify others. Other available commands: start, help, settings")

	start := model.NewAutocompleteData(skyroomStartCommand, "", "Start a new meeting for the current channel")
	skyroom.AddCommand(start)

	help := model.NewAutocompleteData("help", "", "Get slash command help")
	skyroom.AddCommand(help)

	return skyroom
}

func (p *Plugin) ExecuteCommand(c *plugin.Context, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	split := strings.Fields(args.Command)
	command := split[0]
	// var parameters []string
	action := ""
	if len(split) > 1 {
		action = split[1]
	}
	// if len(split) > 2 {
	// 	parameters = split[2:]
	// }

	if command != "/"+skyroomCommand {
		return &model.CommandResponse{}, nil
	}

	switch action {
	case "help":
		return p.executeHelpCommand(c, args)

	// case "settings":
	// 	return p.executeSettingsCommand(c, args, parameters)

	case skyroomStartCommand:
		fallthrough
	default:
		return p.executeStartMeetingCommand(c, args)
	}
}

func startMeetingError(channelID string, detailedError string) (*model.CommandResponse, *model.AppError) {
	return &model.CommandResponse{
			ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
			ChannelId:    channelID,
			Text:         "We could not start a meeting at this time.",
		}, &model.AppError{
			Message:       "We could not start a meeting at this time.",
			DetailedError: detailedError,
		}
}

func (p *Plugin) executeStartMeetingCommand(c *plugin.Context, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	// input := strings.TrimSpace(strings.TrimPrefix(args.Command, "/"+skyroomCommand))
	// input = strings.TrimSpace(strings.TrimPrefix(input, skyroomStartCommand))

	user, appErr := p.API.GetUser(args.UserId)
	if appErr != nil {
		return startMeetingError(args.ChannelId, fmt.Sprintf("getUser() threw error: %s", appErr))
	}

	channel, appErr := p.API.GetChannel(args.ChannelId)
	if appErr != nil {
		return startMeetingError(args.ChannelId, fmt.Sprintf("getChannel() threw error: %s", appErr))
	}

	// userConfig, err := p.getUserConfig(args.UserId)
	// if err != nil {
	// 	return startMeetingError(args.ChannelId, fmt.Sprintf("getChannel() threw error: %s", err))
	// }

	// if userConfig.NamingScheme == skyroomNameSchemeAsk && input == "" {
	// 	if err := p.askMeetingType(user, channel, args.RootId); err != nil {
	// 		return startMeetingError(args.ChannelId, fmt.Sprintf("startMeeting() threw error: %s", appErr))
	// 	}
	// } else {
	// 	if _, err := p.startMeeting(user, channel, "", input, false, args.RootId); err != nil {
	// 		return startMeetingError(args.ChannelId, fmt.Sprintf("startMeeting() threw error: %s", appErr))
	// 	}
	// }

	if _, err := p.startMeeting(user, channel, args.RootId); err != nil {
		return startMeetingError(args.ChannelId, fmt.Sprintf("startMeeting() threw error: %s", appErr))
	}
	return &model.CommandResponse{}, nil
}

func (p *Plugin) executeHelpCommand(c *plugin.Context, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	l := p.b.GetUserLocalizer(args.UserId)
	helpTitle := p.b.LocalizeWithConfig(l, &i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID: "skyroom.command.help.title",
			Other: `###### Mattermost Skyroom Plugin - Slash Command help
`,
		},
	})
	commandHelp := p.b.LocalizeWithConfig(l, &i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID: "skyroom.command.help.text",
			Other: `* |/skyroom| - Create a new meeting and invite others
* |/skyroom start| - Create a new meeting and invite others
* |/skyroom help| - Show this help text`,
		},
	})

	text := helpTitle + strings.ReplaceAll(commandHelp, "|", "`")
	post := &model.Post{
		UserId:    p.botID,
		ChannelId: args.ChannelId,
		Message:   text,
		RootId:    args.RootId,
	}
	_ = p.API.SendEphemeralPost(args.UserId, post)

	return &model.CommandResponse{}, nil
}
