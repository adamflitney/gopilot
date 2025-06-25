package slackmessenger

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"text/template"

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

func (sm *SlackMessenger) SendBlockDirectMessage(userID string, filename string, data map[string]string) error {
	fmt.Printf("user: %s is being sent a message", userID)
	convoParams := slack.OpenConversationParameters{
		Users: []string{userID},
	}
	channel, _, _, _ := sm.client.OpenConversation(&convoParams)
	return sm.SendMessage(channel.ID, filename, data)
}

// send a message to the channel where the command was triggered
func (sm *SlackMessenger) SendMessage(channelID string, filename string, data map[string]string) error {
	var messageBlock slack.Blocks
	byteValue, readError := os.ReadFile(filename)
	if readError != nil {
		return fmt.Errorf("failed reading file %s: %w", filename, readError)
	}

	tmpl, err := template.New("challenge").Parse(string(byteValue))
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	err = json.Unmarshal(buf.Bytes(), &messageBlock)
	if err != nil {
		return fmt.Errorf("failed to unmarshal message block: %w", err)
	}
	_, _, err = sm.client.PostMessage(channelID, slack.MsgOptionBlocks(messageBlock.BlockSet...))
	if err != nil {
		return fmt.Errorf("failed posting message: %v", err)
	}
	return nil
}
