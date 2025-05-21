package rockpaperscissors

import "github.com/slack-go/slack"

// a struct to hold the game state and the methods to play the game
type RockPaperScissors struct {
	// players
	// responses
	// originating channel ID
	messenger *main.SlackMessenger
}

// a method to start the game
func (r *RockPaperScissors) StartGame(api *slack.Client, challenger string, challengee string) {
	// send message to each player

}

func (r *RockPaperScissors) SaveResponse(userId string, response string) {
}
