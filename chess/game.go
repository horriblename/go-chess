package chess

import (
	"errors"
	"strconv"
	"strings"
)

var (
	ErrOutOfTurn       = errors.New("player not in turn")
	ErrIllegalMove     = errors.New("illegal move")
	ErrInvalidUnit     = errors.New("no unit at coordinate")
	ErrInvalidNotation = errors.New("invalid position notation")
)

type ChessPiece string

// TODO: use int and impl MarshallJSON manually
const (
	Pawn   ChessPiece = "pawn"
	Bishop ChessPiece = "bishop"
	Knight ChessPiece = "knight"
	Rook   ChessPiece = "rook"
	Queen  ChessPiece = "queen"
	King   ChessPiece = "king"
)

type Coord struct {
	X, Y int
}

type Player bool

const (
	White Player = false
	Black Player = true
)

type Game struct {
	board Board
	turn  Player
}

// The board is orientated such that Board[0][0] is the white side,
// and Board[7][0] is the black side
type Board [8][8]Cell

type Cell struct {
	Unit *Unit
}

type Unit struct {
	Piece ChessPiece
	Color Player
}

func NewGame() *Game {
	return &Game{
		// TODO: un-reference return value of NewBoard
		board: *NewBoard(),
		turn:  White,
	}
}

func NewBoard() *Board {
	board := Board{}

	for i := 0; i < 8; i++ {
		board[1][i] = Cell{newUnit(Pawn, White)}
		board[6][i] = Cell{newUnit(Pawn, Black)}
	}

	for i, player := range []Player{White, Black} {
		row := i * 7
		board[row][0] = Cell{newUnit(Rook, player)}
		board[row][1] = Cell{newUnit(Knight, player)}
		board[row][2] = Cell{newUnit(Bishop, player)}
		board[row][3] = Cell{newUnit(Queen, player)}
		board[row][4] = Cell{newUnit(King, player)}
		board[row][5] = Cell{newUnit(Bishop, player)}
		board[row][6] = Cell{newUnit(Knight, player)}
		board[row][7] = Cell{newUnit(Rook, player)}
	}

	// TODO: setup board
	return &board
}

func (self *Game) Play(player Player, from Coord, to Coord) error {
	if self.turn != player {
		return ErrOutOfTurn
	}

	legalMoves, err := self.board.legalMoves(from)
	if err != nil {
		return err
	}

	for _, move := range legalMoves {
		if move == to {
			self.board.getCoord(to).Unit = self.board.getCoord(from).Unit
			self.board.getCoord(from).Unit = nil
			if self.turn == White {
				self.turn = Black
			} else {
				self.turn = White
			}
			return nil
		}
	}

	return ErrIllegalMove
}

func (self *Game) Turn() Player {
	return self.turn
}

// returns ErrInvalidUnit if the cell at the given coord is empty
func (self *Board) legalMoves(coord Coord) ([]Coord, error) {
	if self[coord.Y][coord.X].Unit == nil {
		return nil, ErrInvalidUnit
	}

	switch self[coord.Y][coord.X].Unit.Piece {
	case Pawn:
		return self.legalMovesPawn(coord), nil
	case Bishop:
		return self.legalMovesInDirection(coord, false, true), nil
	case Knight:
		return self.legalMovesKnight(coord), nil
	case Rook:
		return self.legalMovesInDirection(coord, true, false), nil
	case Queen:
		return self.legalMovesInDirection(coord, true, true), nil
	case King:
		return self.legalMovesKing(coord), nil
	}

	// unreachable
	return nil, ErrInvalidUnit
}

func (self *Board) legalMovesPawn(coord Coord) []Coord {
	unit := self[coord.Y][coord.X].Unit
	coords := make([]Coord, 0)
	direction := 1
	if unit.Color == Black {
		direction = -1
	}
	var unit_i *Unit
	coords = append(coords, Coord{coord.X, coord.Y + direction})
	for _, ci := range [2]Coord{coord.Add(Coord{-1, 1}), coord.Add(Coord{1, 1})} {
		if !inBounds(ci.X, ci.Y) {
			continue
		}

		unit_i = self.getCoord(ci).Unit
		if unit_i != nil && unit_i.Color != unit.Color {
			coords = append(coords, ci)
		}
	}

	return coords
}

func (self *Board) legalMovesKing(coord Coord) []Coord {
	mover := self.getCoord(coord)
	coords := make([]Coord, 0)
	for _, delta := range []Coord{{-1, -1}, {-1, 0}, {-1, 1}, {0, -1}, {0, 1}, {1, -1}, {1, 0}, {1, 1}} {
		xi := coord.X + delta.X
		yi := coord.Y + delta.Y
		if !inBounds(xi, yi) {
			continue
		}
		cell := self[yi][xi]
		if cell.Unit == nil || cell.Unit.Color != mover.Unit.Color {
			coords = append(coords, Coord{xi, yi})
		}
	}

	return coords
}

func (self *Board) legalMovesKnight(coord Coord) []Coord {
	mover := self.getCoord(coord)
	coords := make([]Coord, 0)
	for _, delta := range []Coord{{-1, 2}, {1, 2}, {2, -1}, {2, 1}, {-1, -2}, {1, -2}, {-2, -1}, {-2, 1}} {
		ci := coord.Add(delta)
		if !inBounds(ci.X, ci.Y) {
			continue
		}
		cell := self.getCoord(ci)
		if cell.Unit == nil || cell.Unit.Color != mover.Unit.Color {
			coords = append(coords, ci)
		}
	}
	return coords
}

// generalizes legal moves of bishop, rook, and queen
// straight indicates if the unit can move in a straight line
// diagonal indicates if the unit can move diagonally
func (self *Board) legalMovesInDirection(coord Coord, straight, diagonal bool) []Coord {
	coords := make([]Coord, 0)

	if straight {
		for _, direction := range []Coord{{-1, 0}, {1, 0}, {0, -1}, {0, 1}} {
			coords = append(coords, self.generateMovesInDirection(coord, direction)...)
		}
	}

	if diagonal {
		for _, direction := range []Coord{{-1, -1}, {-1, 1}, {1, -1}, {1, 1}} {
			coords = append(coords, self.generateMovesInDirection(coord, direction)...)
		}
	}

	return coords
}

func (self *Board) generateMovesInDirection(origin, delta Coord) []Coord {
	coords := make([]Coord, 0)
	for ci := origin.Add(delta); ; ci = ci.Add(delta) {
		if !inBounds(ci.X, ci.Y) {
			break
		}
		unit_i := self.getCoord(ci)
		if unit_i != nil {
			if unit_i.Unit.Color != self.getCoord(origin).Unit.Color {
				coords = append(coords, ci)
			}
			break
		}

		coords = append(coords, ci)
	}
	return coords
}

func newUnit(piece ChessPiece, color Player) *Unit {
	return &Unit{piece, color}
}

func (self *Board) getCoord(coord Coord) *Cell {
	return &self[coord.Y][coord.X]
}

func inBounds(x, y int) bool {
	return x >= 0 && y >= 0 && x < 8 && y < 8
}

func (self Coord) Add(other Coord) Coord {
	return Coord{self.X + other.X, self.Y + other.Y}
}

// converts a position in chess notation into Coord e.g. "a1" -> (0, 0)
func CoordFromChessNotation(pos string) (coord Coord, err error) {
	if len(pos) != 2 {
		return coord, ErrInvalidNotation
	}

	x := int(pos[0] - 'a')
	y := int(pos[1] - '1')

	if !inBounds(x, y) {
		return coord, ErrInvalidNotation
	}

	return Coord{x, y}, nil
}

func (self *Board) Visualize(side Player) string {
	var builder strings.Builder
	builder.WriteString("\n  │ " + strings.Repeat("─", 29) + " │\n")
	for i := range self {
		if side == White {
			i = len(self) - i - 1
		}
		row := &self[i]
		builder.WriteString(strconv.Itoa(i + 1))
		builder.WriteString(" │ ")
		for _, cell := range row {
			if cell.Unit != nil {
				builder.WriteString(cell.Unit.visualize())
			} else {
				builder.WriteString(" ")
			}
			builder.WriteString(" │ ")
		}
		builder.WriteString("\n  │ " + strings.Repeat("─", 29) + " │\n")
	}

	builder.WriteString(" ")
	for i := 'a'; i <= 'h'; i++ {
		builder.WriteString("   " + string(i))
	}
	return builder.String()
}

func (self Unit) visualize() string {
	var s string
	switch self.Piece {
	case Pawn:
		s = "o"
	case Bishop:
		s = "♠"
	case Knight:
		s = "♞"
	case Rook:
		s = "Ψ"
	case Queen:
		s = "Q"
	case King:
		s = "K"
	}
	if self.Color == Black {
		return "\x1b[31m" + s + "\x1b[0m"
	}
	return s
}
