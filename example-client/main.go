package main

import (
	"fmt"
	"log"

	"github.com/horriblename/go-chess/chess"
	proto "github.com/horriblename/go-chess/protocol"
	"golang.org/x/net/websocket"
)

type AppState struct {
	curr *chess.Board
	// allows undoing 1 move (by our side), used to undo illegal moves
	save *chess.Board

	color chess.Player
}

func main() {
	origin := "http://localhost/"
	url := "ws://localhost:9990/game"

	ws, err := websocket.Dial(url, "", origin)
	if err != nil {
		log.Fatal(err)
	}

	appState := AppState{
		curr: chess.NewBoard(),
		save: chess.NewBoard()}

	println("Waiting for game...")

	var data proto.Event
	err = websocket.JSON.Receive(ws, &data)
	if err != nil {
		log.Fatal("receiving json: ", err)
	}

	if data.Message != proto.GameStart {
		log.Fatal("expecting message 'gameStart', got ", data.Message)
	}

	println("Game Started!\n")

	if *data.StartFirst {
		appState.color = chess.White
		appState.playerMove(ws)
	} else {
		appState.color = chess.Black
	}

gameLoop:
	for {
		println(appState.curr.Visualize(appState.color))
		println("Waiting for opponent...")
		err = websocket.JSON.Receive(ws, &data)
		if err != nil {
			log.Fatal("receiving json: ", err)
		}

		switch data.Message {
		case proto.PlayerTurn:
			appState.playMove(data.OpponentMove[0], data.OpponentMove[1])
			fmt.Printf("=== Opponent played: %s -> %s\n", data.OpponentMove[0], data.OpponentMove[1])
		case proto.IllegalMove:
			println("That move was illegal: Please try again")
			appState.rollback()
		case proto.MoveAccepted:
			println("Waiting for opponent...")
		case proto.GameEnded:
			if *data.Winner == "player" {
				println("Game Ended: You Won!")
			} else {
				println("Game Ended: You Lost!")
			}
			break gameLoop

		default:
			log.Printf("unexpected message: %s", data.Message)
		}

		appState.playerMove(ws)
	}
}

func (self *AppState) playerMove(ws *websocket.Conn) {
	from, to := self.promptNextMove()
	move := proto.Request{Request: proto.Move, Move: [2]string{from, to}}
	err := websocket.JSON.Send(ws, move)
	if err != nil {
		log.Printf("sending JSON: %s", err)
	}

	fmt.Printf("=== You played: %s -> %s\n", from, to)
	err = self.playMove(from, to)
	if err != nil {
		panic(err)
	}
}

func (self *AppState) promptNextMove() (from, to string) {
	var err error
	for {
		println(self.curr.Visualize(self.color))
		print("what's your next move? ")
		fmt.Scanf("%s %s", &from, &to)

		_, err = chess.CoordFromChessNotation(from)
		if err != nil {
			fmt.Printf(`unkown position: %s; type "a1 b3" to move the unit on "a1" to "b3"`, from)
			continue
		}
		_, err = chess.CoordFromChessNotation(to)
		if err != nil {
			fmt.Printf(`unkown position: %s; type "a1 b3" to move the unit on "a1" to "b3"`, to)
			continue
		}
		return from, to
	}
}

func (self *AppState) playMove(from string, to string) error {
	*self.save = *self.curr
	var err error
	a, err := chess.CoordFromChessNotation(from)
	if err != nil {
		return err
	}
	b, err := chess.CoordFromChessNotation(to)
	if err != nil {
		return err
	}
	if a == b {
		return fmt.Errorf("moving piece from same place to itself")
	}

	self.curr[b.Y][b.X].Unit = self.curr[a.Y][a.X].Unit
	self.curr[a.Y][a.X].Unit = nil

	return nil
}

func (self *AppState) rollback() {
	*self.curr = *self.save
}
