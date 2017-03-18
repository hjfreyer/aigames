package main

import (
	"log"
	"time"
)

var spaceToQuads [][]int

func init() {
	res := make([][]int, 42)
	var count int
	// Horizontal
	for row := 0; row < 6; row++ {
		for col := 0; col < 4; col++ {
			for pos := 0; pos < 4; pos++ {
				idx := row*7 + col + pos
				res[idx] = append(res[idx], count)
			}
			count++
		}
	}
	// Vertical
	for row := 0; row < 3; row++ {
		for col := 0; col < 7; col++ {
			for pos := 0; pos < 4; pos++ {
				idx := (row+pos)*7 + col
				res[idx] = append(res[idx], count)
			}
			count++
		}
	}
	// Diagonal
	for row := 0; row < 3; row++ {
		for col := 0; col < 4; col++ {
			// Backslash
			for pos := 0; pos < 4; pos++ {
				idx := (row+pos)*7 + col + pos
				res[idx] = append(res[idx], count)
			}
			count++
			// Forwardslash
			for pos := 0; pos < 4; pos++ {
				idx := (row+pos)*7 + col + 3 - pos
				res[idx] = append(res[idx], count)
			}
			count++
		}
	}
	if count != 69 {
		log.Fatal("Bad count!", count)
	}
	spaceToQuads = res
}

type GameEnv struct {
	PlayerID        byte
	G               *gameState
	AlreadyWon      bool
	AlreadyWonDepth int
	RoundNum        int
}

func (e *GameEnv) Reset(field []byte) {
	e.G = NewGame(field, e.PlayerID)
}

type Move struct {
	Col int
	Val float64
}

func (e *GameEnv) NextMove(dlt time.Time) (Move, int) {
	// If we know we've won, do a full search with a trivial window:
	if e.AlreadyWon {
		if e.PlayerID == '1' {
			move, _ := minimax(e.G, e.AlreadyWonDepth, -Inf, -1e100, nil)
			return move, e.AlreadyWonDepth
		}
		move, _ := minimax(e.G, e.AlreadyWonDepth, 1e100, Inf, nil)
		return move, e.AlreadyWonDepth
	}

	depth := 1
	var bestMove Move
	var t ApproximateTimer
	t.Start(dlt)
	for {
		move, cancelled := minimax(e.G, depth, -Inf, Inf, &t)
		if cancelled {
			break
		}

		bestMove = move

		if bestMove.Val == Inf || bestMove.Val == -Inf {
			break
		}

		depth++
	}
	log.Print("Reached depth: ", depth)

	return bestMove, depth
}
