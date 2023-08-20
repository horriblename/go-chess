package chess

import (
	"fmt"
	"testing"
)

func TestMovesPawn(t *testing.T) {
	var board Board
	board[5][5] = Cell{newUnit(Pawn, White)}
	board[6][4] = Cell{newUnit(Pawn, Black)}

	println("hello")
	println(board.Visualize(White))

	x, err := CoordFromChessNotation("a1")
	fmt.Printf("%+v, err: %s\n", x, err)
	x, err = CoordFromChessNotation("h8")
	fmt.Printf("%+v, err: %s\n", x, err)
}
