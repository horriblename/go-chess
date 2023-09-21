package protocol

// idk if there is proper terminology for this, but "Event" means server to client,
// "Request" means client to server

type Event struct {
	Message      EventMessage `json:"message"`
	StartFirst   bool         `json:"startFirst,omitempty"`
	OpponentMove [2]string    `json:"opponentMove,omitempty"`
	Winner       string       `json:"winner,omitempty"`
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
