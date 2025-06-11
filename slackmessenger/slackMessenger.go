package slackmessenger

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
)

type SlackMessenger struct {
	client *socketmode.Client
}

func NewSlackMessenger(client *socketmode.Client) *SlackMessenger {
	return &SlackMessenger{
		client: client,
	}
}

func (sm *SlackMessenger) SendBlockDirectMessage(userID string, filename string) error {
	fmt.Printf("user: %s is being sent a message", userID)
	convoParams := slack.OpenConversationParameters{
		Users: []string{userID},
	}
	channel, _, _, _ := sm.client.OpenConversation(&convoParams)
	sm.SendMessage(channel.ID, filename)
	return nil
}

// send a message to the channel where the command was triggered
func (sm *SlackMessenger) SendMessage(channelID string, filename string) error {
	var messageBlock slack.Blocks
	byteValue, readError := os.ReadFile(filename)
	if readError != nil {
		return fmt.Errorf("failed reading file %s: %w", filename, readError)
	}
	json.Unmarshal(byteValue, &messageBlock)
	_, _, err := sm.client.PostMessage(channelID, slack.MsgOptionBlocks(messageBlock.BlockSet...))
	if err != nil {
		return fmt.Errorf("failed posting message: %v", err)
	}
	return nil
}
