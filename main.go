package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/slack-go/slack/socketmode"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

func main() {
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

	go func() {
		for evt := range client.Events {
			switch evt.Type {
			case socketmode.EventTypeConnecting:
				fmt.Println("Connecting to Slack with Socket Mode...")
			case socketmode.EventTypeConnectionError:
				fmt.Println("Connection failed. Retrying later...")
			case socketmode.EventTypeConnected:
				fmt.Println("Connected to Slack with Socket Mode.")
			case socketmode.EventTypeEventsAPI:
				eventsAPIEvent, ok := evt.Data.(slackevents.EventsAPIEvent)
				if !ok {
					fmt.Printf("Ignored %+v\n", evt)

					continue
				}

				fmt.Printf("Event received: %+v\n", eventsAPIEvent)

				client.Ack(*evt.Request)

				switch eventsAPIEvent.Type {
				case slackevents.CallbackEvent:
					innerEvent := eventsAPIEvent.InnerEvent
					switch ev := innerEvent.Data.(type) {
					case *slackevents.AppMentionEvent:
						_, _, err := client.PostMessage(ev.Channel, slack.MsgOptionText("Yes, hello.", false))
						if err != nil {
							fmt.Printf("failed posting message: %v", err)
						}
					case *slackevents.MemberJoinedChannelEvent:
						fmt.Printf("user %q joined to channel %q", ev.User, ev.Channel)
					}
				default:
					client.Debugf("unsupported Events API event received")
				}
			case socketmode.EventTypeInteractive:
				callback, ok := evt.Data.(slack.InteractionCallback)
				if !ok {
					fmt.Printf("Ignored %+v\n", evt)

					continue
				}

				fmt.Printf("Interaction received: %+v\n", callback)

				var payload interface{}

				switch callback.Type {
				case slack.InteractionTypeBlockActions:
					// See https://api.slack.com/apis/connections/socket-implement#button
				// get the button action from the callback
					actionCallback := callback.ActionCallback.BlockActions[0]
					// get the user name from the callback
					userName := callback.User.Name

					client.Debugf("button clicked: %s, %s", actionCallback.Value, userName)
					payload = map[string]interface{}{
						"text": "thank you world",
					}
					client.Ack(*evt.Request, payload)
					// update the interacted message to replace the buttons with text
					_, _, err := client.PostMessage(callback.Channel.ID, slack.MsgOptionText(fmt.Sprintf("Thank you %s!", userName), false), slack.MsgOptionReplaceOriginal(callback.ResponseURL))
					if err != nil {
						fmt.Printf("failed posting message: %v", err)
					}

				case slack.InteractionTypeShortcut:
				case slack.InteractionTypeViewSubmission:
					// See https://api.slack.com/apis/connections/socket-implement#modal
				case slack.InteractionTypeDialogSubmission:
				default:

				}
				

				client.Ack(*evt.Request, payload)
			case socketmode.EventTypeSlashCommand:
				cmd, ok := evt.Data.(slack.SlashCommand)
				if !ok {
					fmt.Printf("Ignored %+v\n", evt)

					continue
				}

				client.Debugf("Slash command received: %+v", cmd)

				payload := map[string]interface{}{
					"text": "hello world",
				}

				client.Ack(*evt.Request, payload)
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
						var user slack.User
						for _, foundUser := range users {
							if foundUser.Name == userName {
								user = foundUser
								break
							}
						}
						_, _, err = client.PostMessage(cmd.ChannelID, slack.MsgOptionText(fmt.Sprintf("Hello %s!", user.Name), false))
						if err != nil {
							fmt.Printf("failed posting message: %v", err)
						}
						sendChallengeDirectMessage(client, user.ID)
						sendChallengeDirectMessage(client, cmd.UserID)	
					} else {
						_, _, err := client.PostMessage(cmd.ChannelID, slack.MsgOptionText(fmt.Sprintf("Hello %s!", cmd.Text), false))
						if err != nil {
							fmt.Printf("failed posting message: %v", err)
						}
					}
				}
				// if arg passed, lookup the user
			case socketmode.EventTypeHello:
				client.Debugf("Hello received!")
			default:
				fmt.Fprintf(os.Stderr, "Unexpected event type received: %s\n", evt.Type)
			}
		}
	}()

	client.Run()
}

func sendChallengeDirectMessage(client *socketmode.Client, userID string) {
	convoParams := slack.OpenConversationParameters{
		Users: []string{userID},
	}
	channel, _, _, _ := client.OpenConversation(&convoParams)
	var challengeBlock slack.Blocks;
	byteValue, _ := os.ReadFile("challenge.json")
	json.Unmarshal(byteValue, &challengeBlock);
	fmt.Printf("challengeBlock: %v", challengeBlock)
	_, _, err := client.PostMessage(channel.ID, slack.MsgOptionBlocks(challengeBlock.BlockSet...))
	if err != nil {
		fmt.Printf("failed posting message: %v", err)
	}
}
