package rockpaperscissors

import (
	"fmt"
	"gopilot/slackmessenger"
)

type Player struct {
	ID     string
	Name   string
	Handle string
}

// a struct to hold the game state and the methods to play the game
type RockPaperScissors struct {
	responses map[Player]string
	channelId string
	messenger *slackmessenger.SlackMessenger
	players   []Player
}

func NewRockPaperScissors(messenger *slackmessenger.SlackMessenger) *RockPaperScissors {
	return &RockPaperScissors{
		players:   make([]Player, 2),
		messenger: messenger,
	}
}

// a method to start the game
func (r *RockPaperScissors) StartGame(channelID string, challenger Player, challengee Player) {
	// send message to each player
	r.channelId = channelID
	r.players = append(r.players, challenger, challengee)

	data := map[string]string{"Name": challenger.Handle}
	err := r.messenger.SendBlockDirectMessage(challengee.ID, "rockpaperscissors/challenge.json", data)
	if err != nil {
		fmt.Printf("failed to send challenge message to %s: %v", challengee.Handle, err)
	}

	err = r.messenger.SendBlockDirectMessage(challenger.ID, "rockpaperscissors/challenge.json", data)
	if err != nil {
		fmt.Printf("failed to send challenge message to %s: %v", challenger.Handle, err)
	}
}

func (r *RockPaperScissors) SaveResponse(userId string, response string) {
	if r.responses == nil {
		r.responses = make(map[Player]string)
	}
	var player *Player
	for i := range r.players {
		if r.players[i].ID == userId {
			player = &r.players[i]
			break
		}
	}
	if player == nil {
		fmt.Printf("Player with ID %s not found", userId)
		return
	}
	r.responses[*player] = response

	if len(r.responses) == 2 {
		r.determineWinner()
	}
}

func (r *RockPaperScissors) determineWinner() {

	// Determine winner based on rock-paper-scissors rules
	if r.responses[r.players[0]] == r.responses[r.players[1]] {
		// It's a tie
		data := map[string]string{"Player1": r.players[0].Handle, "Player2": r.players[1].Handle}
		r.messenger.SendMessage(r.channelId, "rockpaperscissors/draw.json", data)
		r.messenger.SendBlockDirectMessage(r.players[0].ID, "rockpaperscissors/draw.json", data)
		r.messenger.SendBlockDirectMessage(r.players[1].ID, "rockpaperscissors/draw.json", data)
		return
	}

	var winner, loser Player
	switch {
	case r.responses[r.players[0]] == "rock" && r.responses[r.players[1]] == "scissors",
		r.responses[r.players[0]] == "scissors" && r.responses[r.players[1]] == "paper",
		r.responses[r.players[0]] == "paper" && r.responses[r.players[1]] == "rock":
		winner = r.players[0]
		loser = r.players[1]
	default:
		winner = r.players[1]
		loser = r.players[0]
	}
	data := map[string]string{"Winner": winner.Handle, "Loser": loser.Handle}
	r.messenger.SendMessage(r.channelId, "rockpaperscissors/result.json", data)
	r.messenger.SendBlockDirectMessage(winner.ID, "rockpaperscissors/winner.json", data)
	r.messenger.SendBlockDirectMessage(loser.ID, "rockpaperscissors/loser.json", data)
}
