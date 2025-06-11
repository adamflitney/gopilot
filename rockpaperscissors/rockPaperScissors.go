package rockpaperscissors

import (
	"fmt"
	"gopilot/slackmessenger"
	"maps"
	"slices"
)

// a struct to hold the game state and the methods to play the game
type RockPaperScissors struct {
	responses map[string]string
	channelId string
	messenger *slackmessenger.SlackMessenger
}

func NewRockPaperScissors(messenger *slackmessenger.SlackMessenger) *RockPaperScissors {
	return &RockPaperScissors{
		// players:     make([]string, 0),
		// responses:   make(map[string]string),
		messenger: messenger,
	}
}

// a method to start the game
func (r *RockPaperScissors) StartGame(channelID string, challenger string, challengee string) {
	// send message to each player
	r.channelId = channelID

	err := r.messenger.SendBlockDirectMessage(challengee, "rockpaperscissors/challenge.json")
	if err != nil {
		fmt.Printf("failed to send challenge message to %s: %v", challengee, err)
	}

	err = r.messenger.SendBlockDirectMessage(challenger, "rockpaperscissors/challenge.json")
	if err != nil {
		fmt.Printf("failed to send challenge message to %s: %v", challengee, err)
	}
}

func (r *RockPaperScissors) SaveResponse(userId string, response string) {
	if r.responses == nil {
		r.responses = make(map[string]string)
	}
	r.responses[userId] = response

	if len(r.responses) == 2 {
		r.determineWinner()
	}
}

func (r *RockPaperScissors) determineWinner() {

	players := slices.Collect(maps.Keys(r.responses))

	// Determine winner based on rock-paper-scissors rules
	if r.responses[players[0]] == r.responses[players[1]] {
		// It's a tie
		r.messenger.SendMessage(r.channelId, "rockpaperscissors/draw.json")
		r.messenger.SendBlockDirectMessage(players[0], "rockpaperscissors/draw.json")
		r.messenger.SendBlockDirectMessage(players[1], "rockpaperscissors/draw.json")
		return
	}

	var winner, loser string
	switch {
	case r.responses[players[0]] == "rock" && r.responses[players[1]] == "scissors",
		r.responses[players[0]] == "scissors" && r.responses[players[1]] == "paper",
		r.responses[players[0]] == "paper" && r.responses[players[1]] == "rock":
		winner = players[0]
		loser = players[1]
	default:
		winner = players[1]
		loser = players[0]
	}
	r.messenger.SendMessage(r.channelId, "rockpaperscissors/winner.json")
	r.messenger.SendBlockDirectMessage(winner, "rockpaperscissors/winner.json")
	r.messenger.SendBlockDirectMessage(loser, "rockpaperscissors/loser.json")
}
