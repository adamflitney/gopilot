package main

import (
	"fmt"
	"gopilot/rockpaperscissors"
	"gopilot/slackmessenger"
	"log"
	"os"
	"strings"

	"github.com/slack-go/slack/socketmode"

	"github.com/slack-go/slack"
)

var game *rockpaperscissors.RockPaperScissors

func main() {
	// connect to slack API
	appToken := os.Getenv("SLACK_APP_TOKEN")
	if appToken == "" {
		fmt.Fprintf(os.Stderr, "SLACK_APP_TOKEN must be set.\n")
		os.Exit(1)
	}

	if !strings.HasPrefix(appToken, "xapp-") {
		fmt.Fprintf(os.Stderr, "SLACK_APP_TOKEN must have the prefix \"xapp-\".")
	}

	botToken := os.Getenv("SLACK_BOT_TOKEN")
	if botToken == "" {
		fmt.Fprintf(os.Stderr, "SLACK_BOT_TOKEN must be set.\n")
		os.Exit(1)
	}

	if !strings.HasPrefix(botToken, "xoxb-") {
		fmt.Fprintf(os.Stderr, "SLACK_BOT_TOKEN must have the prefix \"xoxb-\".")
	}

	api := slack.New(
		botToken,
		slack.OptionDebug(true),
		slack.OptionLog(log.New(os.Stdout, "api: ", log.Lshortfile|log.LstdFlags)),
		slack.OptionAppLevelToken(appToken),
	)

	client := socketmode.New(
		api,
		socketmode.OptionDebug(true),
		socketmode.OptionLog(log.New(os.Stdout, "socketmode: ", log.Lshortfile|log.LstdFlags)),
	)

	// listen to events from slack client
	go func() {
		for evt := range client.Events {
			switch evt.Type {
			case socketmode.EventTypeEventsAPI:
				client.Ack(*evt.Request)
			case socketmode.EventTypeInteractive:
				processInteractionEvent(client, evt)
			case socketmode.EventTypeSlashCommand:
				processSlashCommand(api, client, evt)
			default:
				fmt.Fprintf(os.Stderr, "Unexpected event type received: %s\n", evt.Type)
			}
		}
	}()

	client.Run()
}

func processInteractionEvent(client *socketmode.Client, evt socketmode.Event) {
	callback, ok := evt.Data.(slack.InteractionCallback)
	if !ok {
		fmt.Printf("Ignored %+v\n", evt)
		return
	}

	fmt.Printf("Interaction received: %+v\n", callback)

	var payload interface{}

	if callback.Type == slack.InteractionTypeBlockActions {
		// See https://api.slack.com/apis/connections/socket-implement#button
		// get the button action from the callback
		actionCallback := callback.ActionCallback.BlockActions[0]
		// get the user name from the callback
		userName := callback.User.Name
		userId := callback.User.ID

		client.Debugf("button clicked: %s, %s", actionCallback.Value, userName)
		payload = map[string]interface{}{
			"text": "thank you world",
		}
		// update the interacted message to replace the buttons with text
		_, _, err := client.PostMessage(callback.Channel.ID, slack.MsgOptionText(fmt.Sprintf("Thank you %s!", userName), false), slack.MsgOptionReplaceOriginal(callback.ResponseURL))
		if err != nil {
			fmt.Printf("failed posting message: %v", err)
		}
		game.SaveResponse(userId, actionCallback.Value)
	}
	client.Ack(*evt.Request, payload)
}

func processSlashCommand(api *slack.Client, client *socketmode.Client, evt socketmode.Event) {
	cmd, ok := evt.Data.(slack.SlashCommand)
	if !ok {
		fmt.Printf("Ignored %+v\n", evt)
		return
	}

	client.Debugf("Slash command received: %+v", cmd)

	payload := map[string]interface{}{
		"text": "hello world",
	}

	client.Ack(*evt.Request, payload)

	messenger := slackmessenger.NewSlackMessenger(client)
	game = rockpaperscissors.NewRockPaperScissors(messenger)

	// check if an argument is passed
	if len(cmd.Text) > 0 {
		// check if the argument is a user
		if strings.HasPrefix(cmd.Text, "@") {
			userName := strings.TrimPrefix(cmd.Text, "@")
			// find the userId from the user name
			users, err := api.GetUsers()
			if err != nil {
				fmt.Printf("failed getting users: %v", err)
				return
			}
			var challengee slack.User
			for _, foundUser := range users {
				if foundUser.Name == userName {
					challengee = foundUser
					break
				}
			}
			challengeePlayer := rockpaperscissors.Player{
				ID:     challengee.ID,
				Handle: challengee.Name,
				Name:   challengee.RealName,
			}
			// loop through users match where .Name == cmd.UserName
			var challenger slack.User
			for _, foundUser := range users {
				if foundUser.Name == cmd.UserName {
					challenger = foundUser
					break
				}
			}
			challengerPlayer := rockpaperscissors.Player{
				ID:     challenger.ID,
				Handle: challenger.Name,
				Name:   challenger.RealName,
			}
			game.StartGame(cmd.ChannelID, challengerPlayer, challengeePlayer)

		} else {
			_, _, err := client.PostMessage(cmd.ChannelID, slack.MsgOptionText(fmt.Sprintf("Hello %s!", cmd.Text), false))
			if err != nil {
				fmt.Printf("failed posting message: %v", err)
			}
		}
	}
}
