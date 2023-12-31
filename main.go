package main

import (
	"errors"
	"io"
	"log"
	"math/rand"
	"net/http"

	"github.com/horriblename/go-chess/chess"
	proto "github.com/horriblename/go-chess/protocol"

	"golang.org/x/net/websocket"
)

type gameID int
type userID int

type idPair struct {
	user userID
	game gameID
}

type gameServer struct {
	userSessions map[userID]gameID
	gameSessions map[gameID]*gameSession
	// waiting      chan userID
	idGenerator chan userID
}

type gameSession struct {
	game         *chess.Game
	white        userID
	black        userID
	whiteReceive chan receiveType
	whiteSend    chan proto.Event
	blackReceive chan receiveType
	blackSend    chan proto.Event
}

// why can't we have union typpppppppppppes
type receiveType struct {
	req          *proto.Request
	disconnected bool
	connected    bool
}

type requestType string

const (
	play requestType = "play"
)

var (
	errNewUserIDTooManyTries = errors.New("too many tries while creating new user id")
)

func newGameServer() *gameServer {
	server := &gameServer{
		make(map[userID]gameID, 0),
		make(map[gameID]*gameSession, 0),
		// XXX: channel size must be >= 2
		make(chan userID, 20),
	}
	go server.gameSessionGenerator()
	return server
}

func (self *gameServer) gameSessionGenerator() {
	for {
		gid := self.generateUniqueGameID()
		n1 := self.generateUniqueUserID()
		n2 := self.generateUniqueUserID()
		session := &gameSession{
			white:        n1,
			black:        n2,
			whiteReceive: make(chan receiveType),
			whiteSend:    make(chan proto.Event),
			blackReceive: make(chan receiveType),
			blackSend:    make(chan proto.Event),
		}
		self.gameSessions[gid] = session

		self.userSessions[n1] = gid
		self.userSessions[n2] = gid

		go session.startGame()
		self.idGenerator <- n1
		self.idGenerator <- n2
	}
}

func (self *gameServer) generateUniqueUserID() userID {
	// TODO: limit loop count and return an error
	for {
		n := userID(rand.Int())
		if _, ok := self.userSessions[n]; !ok {
			return n
		}
	}
}
func (self *gameServer) generateUniqueGameID() gameID {
	// TODO: limit loop count and return an error
	for {
		n := gameID(rand.Int())
		if _, ok := self.gameSessions[n]; !ok {
			return n
		}
	}
}

func (self *gameServer) handleConnection(ws *websocket.Conn) {
	var err error
	var data proto.Request
	userID := <-self.idGenerator
	gameID := self.userSessions[userID]
	gameSession := self.gameSessions[gameID]
	var requests chan receiveType
	var events chan proto.Event

	if gameSession.white == userID {
		requests = gameSession.whiteReceive
		events = gameSession.whiteSend
	} else {
		requests = gameSession.blackReceive
		events = gameSession.blackSend
	}
	defer func() {
		requests <- receiveType{disconnected: true}
		close(requests)
	}()

	requests <- receiveType{connected: true}

	go func() {
		var ev proto.Event
		var more bool
		for {
			ev, more = <-events
			if !more {
				break
			}
			if err = websocket.JSON.Send(ws, &ev); err != nil {
				log.Println("FIXME: error handling ", err)
			}
		}
	}()

	for {
		if err = websocket.JSON.Receive(ws, &data); err != nil {
			if err == io.EOF {
				break
			}
			log.Println("FIXME: error handling ", err)
			break
		}

		data_copy := data
		requests <- receiveType{req: &data_copy}
	}

	log.Println("closed one user session #", userID)
}

func (self *gameSession) startGame() {
	var err error
	var from, to chess.Coord
	// wait for players
	for ready1, ready2 := false, false; !ready1 || !ready2; {
		select {
		case rcv, more := <-self.whiteReceive:
			if !more || rcv.disconnected {
				close(self.whiteSend)
			}
			ready1 = true
		case rcv, more := <-self.blackReceive:
			if !more || rcv.disconnected {
				close(self.blackSend)
			}
			ready2 = true
		}
	}
	defer close(self.whiteSend)
	defer close(self.blackSend)

	self.game = chess.NewGame()

	self.whiteSend <- proto.Event{Message: proto.GameStart, StartFirst: &proto.True}
	self.blackSend <- proto.Event{Message: proto.GameStart, StartFirst: &proto.False}

	var req receiveType
	var sendPlayer chan proto.Event
	var sendOpponent chan proto.Event
	var player chess.Player
	sendIllegalMove := func() {
		sendPlayer <- proto.Event{
			Message: proto.IllegalMove,
		}
	}

gameLoop:
	for {
		select {
		case rcv, more := <-self.whiteReceive:
			if !more || rcv.disconnected {
				break gameLoop
			}
			if self.game.Turn() != chess.White {
				// ignore messages that are out of turn
				continue
			}
			req = rcv
			sendPlayer = self.whiteSend
			sendOpponent = self.blackSend
			player = chess.White

		case rcv, more := <-self.blackReceive:
			if !more || rcv.disconnected {
				break gameLoop
			}
			if self.game.Turn() != chess.Black {
				// ignore messages that are out of turn
				continue
			}
			req = rcv
			sendPlayer = self.blackSend
			sendOpponent = self.whiteSend
			player = chess.Black
		}

		if req.req == nil || req.req.Request != proto.Move {
			sendIllegalMove()
			continue
		}

		from, err = chess.CoordFromChessNotation(req.req.Move[0])
		if err != nil {
			sendIllegalMove()
			continue
		}
		to, err = chess.CoordFromChessNotation(req.req.Move[1])
		if err != nil {
			sendIllegalMove()
			continue
		}
		check, err := self.game.Play(player, from, to)
		if err != nil {
			sendIllegalMove()
			continue
		}

		sendPlayer <- proto.Event{
			Message: proto.MoveAccepted,
			Check:   check,
		}

		sendOpponent <- proto.Event{
			Message:      proto.PlayerTurn,
			OpponentMove: &req.req.Move,
			Check:        check,
		}

		if check == chess.CheckMate {
			sendPlayer <- proto.Event{
				Message: proto.GameEnded,
				Winner:  &proto.Player,
			}
			sendOpponent <- proto.Event{
				Message: proto.GameEnded,
				Winner:  &proto.Opponent,
			}
			break gameLoop
		}
	}

	log.Println("game session ended")
}

func main() {
	mux := http.NewServeMux()
	state := newGameServer()

	mux.Handle("/game", websocket.Handler(state.handleConnection))
	mux.Handle("/", http.FileServer(http.Dir("./pages")))

	server := http.Server{
		Handler: mux,
		Addr:    ":9990",
	}

	println("server started")
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("ListenAndServe: %s\n", err)
	}

}
