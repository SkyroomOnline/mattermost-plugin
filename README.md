# Mattermost Skyroom Plugin (Beta)

Create a Skyroom Conferencing room for your current channel in Mattermost, and invite others in that channel to join you via Skyroom's invitations link.

Clicking a video icon in a Mattermost channel posts a message that invites team members to join a Skyroom meetings call.

You can all so use `/skyroom` command to start a new meeting too.

## Installation

Make the project using make file and then add the release to Mattermost plugins folder.

## Configuration

Go to **System Console > Plugins > Skyroom.Online** and set the following values:

1. **Enable Plugin**: ``true``
2. **Skyroom Server URL**: The URL for Skyroom service. https://skyroom.online for example.
3. **Your Skyroom API Key**: The api key you have aquired to call Skyroom's webservice api.
4. **Skyroom Join Link Valid Duration** in minutes. Defaults to 30 minutes.

5. **Encryption Key**: 16 character length alphanumeric string to be used as encryption key for AES encryption algorithm.

You're all set! To test it, go to any Mattermost channel and click the video icon in the channel header to start a new Skyroom meeting.

## Localization

### Localization

Mattermost Skyroom Plugin supports localization in these languages at the moment:
- English
- Farsi

The plugin automatically displays languages based on the following:
- For system messages, the locale set in **System Console > General > Localization > Default Server Language** is used.
- For user messages, such as help text and error messages, the locale set set in **Account Settings > Display > Language** is used.

### Manual builds

You can use build server and webapp projects separately using their manual and setup all together as a mattermost plugin too.

see the `Makefile` for more details or read the `Developing` section.

## Developing

This plugin contains both a server and web app portion.

Use `make` to check the quality of your code, as well as build distributions of the plugin that you can upload to a Mattermost server for testing.

### Server

Inside the `/server` directory, you will find the Go files that make up the server-side of the plugin. Within there, build the plugin like you would any other Go application.

### Web App

Inside the `/webapp` directory, you will find the JS and React files that make up the client-side of the plugin. Within there, modify files and components as necessary. Test your syntax by running `npm run build`.
