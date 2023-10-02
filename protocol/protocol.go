package protocol

import "github.com/horriblename/go-chess/chess"

// idk if there is proper terminology for this, but "Event" means server to client,
// "Request" means client to server

type Winner string

var (
	// to easily make *bool values
	True  bool = true
	False bool = false

	Player   Winner = "player"
	Opponent Winner = "opponent"
)

type Event struct {
	Message      EventMessage      `json:"message"`
	StartFirst   *bool             `json:"startFirst,omitempty"`
	OpponentMove *[2]string        `json:"opponentMove,omitempty"`
	Winner       *Winner           `json:"winner,omitempty"`
	Check        chess.CheckStatus `json:"check,omitempty"`
}

type Request struct {
	Request RequestType `json:"request"`
	Move    [2]string   `json:"move,omitempty"`
}

type EventMessage string

const (
	GameStart    EventMessage = "gameStart"
	PlayerTurn   EventMessage = "playerTurn"
	IllegalMove  EventMessage = "illegalMove"
	MoveAccepted EventMessage = "moveAccepted"
	GameEnded    EventMessage = "gameEnded"
)

type RequestType string

const (
	Move RequestType = "move"
)
