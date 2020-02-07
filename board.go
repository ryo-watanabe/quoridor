package main

import (
	"fmt"
)

type position struct {
	X int	`json:"x"`
	Y int	`json:"y"`
}

func (p *position)Equals(pos position) bool {
	return (p.X == pos.X && p.Y == pos.Y)
}

type wall struct {
	Pole position	`json:"poles"`
	Blockings [][]position	`json:"blockings"`
}

type QuoridorBoard struct {
	Dimension int	`json:"dimension"`
	Poles []position	`json:"poles"`
	Blockings [][]position	`json:"blockings"`
	PlayerPos position	`json:"playerPos"`
	ComPos position	`json:"comPos"`
	PlayerWalls int `json:"playerWalls"`
	ComWalls int `json:"comWalls"`
}

func (b *QuoridorBoard)Copy() *QuoridorBoard {
	var board QuoridorBoard
	board = *b
	board.Poles = make([]position,0)
	board.Poles = append(board.Poles, b.Poles...)
	board.Blockings = make([][]position,0)
	for _, block := range(b.Blockings) {
		newBlock := make([]position,0)
		newBlock = append(newBlock, block...)
		board.Blockings = append(board.Blockings, newBlock)
	}
	return &board
}

func maxRoute(board *QuoridorBoard) int {
	return board.Dimension*board.Dimension
}

func isBlocked(a, b position, board *QuoridorBoard) bool {
	for _, block := range(board.Blockings) {
		if (a.Equals(block[0]) && b.Equals(block[1])) ||
		   (a.Equals(block[1]) && b.Equals(block[0])) {
			return true
		}
	}
	return false
}

func isOnBoard(pos position, board *QuoridorBoard) bool {
	if pos.X < 0 || pos.X >= board.Dimension || pos.Y < 0 || pos.Y >= board.Dimension {
		return false
	}
	return true
}

func possibleMovesFrom(pos position, board *QuoridorBoard) []position {
	moves := make([]position, 0)
	// Right
	moveright := position{X:pos.X+1, Y:pos.Y}
	if moveright.X < board.Dimension && !isBlocked(pos, moveright, board) {
		moves = append(moves, moveright)
	}
	// Left
	moveleft := position{X:pos.X-1, Y:pos.Y}
	if moveleft.X >= 0 && !isBlocked(pos, moveleft, board) {
		moves = append(moves, moveleft)
	}
	// Up
	moveup := position{X:pos.X, Y:pos.Y-1}
	if moveup.Y >= 0 && !isBlocked(pos, moveup, board) {
		moves = append(moves, moveup)
	}
	// Down
	movedown := position{X:pos.X, Y:pos.Y+1}
	if movedown.Y < board.Dimension && !isBlocked(pos, movedown, board) {
		moves = append(moves, movedown)
	}
	return moves
}

func possibleComMoves(board *QuoridorBoard) []position {
	moves := make([]position, 0)
	comMoves := possibleMovesFrom(board.ComPos, board)
	for _, m := range(comMoves) {
		// Jump over player pos
		if m.Equals(board.PlayerPos) {
			playerMoves := possibleMovesFrom(m, board)
			for _, pm := range(playerMoves) {
				if !pm.Equals(board.ComPos) {
					moves = append(moves, pm)
				}
			}
		} else {
			moves = append(moves, m)
		}
	}
	return moves
}

func possiblePlayerMoves(board *QuoridorBoard) []position {
	moves := make([]position, 0)
	playerMoves := possibleMovesFrom(board.PlayerPos, board)
	for _, m := range(playerMoves) {
		// Jump over com pos
		if m.Equals(board.ComPos) {
			comMoves := possibleMovesFrom(m, board)
			for _, pm := range(comMoves) {
				if !pm.Equals(board.PlayerPos) {
					moves = append(moves, pm)
				}
			}
		} else {
			moves = append(moves, m)
		}
	}
	return moves
}

func isPossible(w *wall, board *QuoridorBoard) bool {
	for _, builtPole := range(board.Poles) {
		if builtPole.Equals(w.Pole) {
			return false
		}
	}
	if isBlocked(w.Blockings[0][0], w.Blockings[0][1], board) {
		return false
	}
	if isBlocked(w.Blockings[1][0], w.Blockings[1][1], board) {
		return false
	}
	return true
}

func possibleWalls(board *QuoridorBoard) []wall {
	walls := make([]wall, 0)
	for i := 0; i < board.Dimension - 1; i++ {
		for j := 0; j < board.Dimension - 1; j++ {
			// vertical
			vw := wall{
				Pole: position{Y:i, X:j},
				Blockings: [][]position{
					{position{Y:i, X:j},position{Y:i, X:j+1}},
					{position{Y:i+1, X:j},position{Y:i+1, X:j+1}},
				},
			}
			if isPossible(&vw, board) {
				walls = append(walls, vw)
			}
			// horizontal
			hw := wall{
				Pole: position{Y:i, X:j},
				Blockings: [][]position{
					{position{Y:i, X:j},position{Y:i+1, X:j}},
					{position{Y:i, X:j+1},position{Y:i+1, X:j+1}},
				},
			}
			if isPossible(&hw, board) {
				walls = append(walls, hw)
			}
		}
	}
	return walls
}

func initWalls(dim int) int {
	switch dim {
	case 5:
		return 4
	case 7:
		return 7
	case 9:
		return 10
	}
	return 0
}

func initBoard(req *QuoridorRequest, ret *QuoridorResponse) error {
	dim := 5
	if req.Board != nil && req.Board.Dimension != 0 {
		if req.Board.Dimension > 4  && req.Board.Dimension < 14 {
			dim = req.Board.Dimension
		} else {
			return fmt.Errorf("Invalid dimension %d", req.Board.Dimension)
		}
	}
	ret.Board = &QuoridorBoard{
		Dimension: dim,
		ComPos: position{Y:0, X:dim/2},
		PlayerPos: position{Y:dim-1, X:dim/2},
		ComWalls: initWalls(dim),
		PlayerWalls: initWalls(dim),
	}
	return nil
}
