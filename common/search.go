package main

import (
	"errors"
	"log"
	"math"
	"time"
)

type Player int

const (
	Player1 Player = 1
	Player2 Player = 2
)

func Opponent(p Player) Player {
	switch p {
	case Player1:
		return Player2
	case Player2:
		return Player1
	default:
		panic("Bad player")
	}
}

var (
	ErrDeadlineExceeded = errors.New("Deadline Exceeded")
)

type Score struct {
	GameOver bool
	Player1  float64
	Player2  float64
}

func (s Score) Diff() float64 {
	return s.Player2 - s.Player1
}

func (s Score) IsWinForPlayer(p Player) bool {
	if !s.GameOver {
		return false
	}
	switch p {
	case Player1:
		return math.IsInf(s.Player1, 1)
	case Player2:
		return math.IsInf(s.Player2, 1)
	default:
		panic("bad player")
	}
}

type Move interface{}
type Patch interface{}

type GameEngine interface {
	ResetGame(state string, toMove Player)

	GuessScore() Score
	CurrentPlayer() Player

	ValidMoves() []Move
	Move(Move) Patch
	Reverse(Patch)
}

type Searcher struct {
	G               GameEngine
	Player          Player
	AlreadyWon      bool
	AlreadyWonDepth int
}

func (s *Searcher) Reset(state string) {
	s.G.ResetGame(state, s.Player)
}

func (e *Searcher) NextMove(deadline time.Time) (Move, Score, error) {
	// If we know we've won, do a full search with a trivial window:
	if e.AlreadyWon {
		if e.Player == Player1 {
			return e.minimax(e.AlreadyWonDepth, math.Inf(-1), -1e100, nil)
		}
		return e.minimax(e.AlreadyWonDepth, 1e100, math.Inf(1), nil)
	}

	depth := 1
	var bestMove Move
	var bestScore Score
	var t ApproximateTimer
	t.Start(deadline)
	for {
		move, score, err := e.minimax(depth, math.Inf(-1), math.Inf(1), &t)
		if err == ErrDeadlineExceeded {
			break
		}
		if err != nil {
			return nil, Score{}, err
		}

		// If the opponent won, fall back to the previous depth's best score in the
		// hopes that they blunder.
		if !score.IsWinForPlayer(Opponent(e.Player)) {
			bestMove = move
			bestScore = score
		}

		if score.GameOver {
			break
		}

		depth++
	}
	log.Print("Reached depth: ", depth)

	if e.isWin(bestScore) {
		e.AlreadyWon = true
		e.AlreadyWonDepth = depth
	}

	return bestMove, bestScore, nil
}

func (s *Searcher) isWin(score Score) bool {
	switch s.Player {
	case Player1:
		return math.IsInf(score.Player1, 1)
	case Player2:
		return math.IsInf(score.Player2, 1)
	default:
		panic("Bad player value")
	}
}

func (s *Searcher) minimax(depth int, alpha, beta float64, t *ApproximateTimer) (Move, Score, error) {
	if t.DeadlineExceeded() {
		return nil, Score{}, ErrDeadlineExceeded
	}

	score := s.G.GuessScore()
	if depth == 0 || score.GameOver {
		return nil, score, nil
	}

	switch s.G.CurrentPlayer() {
	case Player2:
		var bestMove Move
		var bestScore Score
		bestDiff := math.Inf(-1)

		for _, move := range s.G.ValidMoves() {
			p := s.G.Move(move)
			_, score, err := s.minimax(depth-1, alpha, beta, t)
			if err != nil {
				return nil, Score{}, err
			}
			s.G.Reverse(p)
			if bestDiff < score.Diff() {
				bestMove = move
				bestScore = score
				bestDiff = score.Diff()
			}
			if alpha < bestDiff {
				alpha = bestDiff
			}
			if beta <= alpha {
				break
			}
		}
		return bestMove, bestScore, nil
	case Player1:
		var bestMove Move
		var bestScore Score
		bestDiff := math.Inf(1)

		for _, move := range s.G.ValidMoves() {
			p := s.G.Move(move)
			_, score, err := s.minimax(depth-1, alpha, beta, t)
			if err != nil {
				return nil, Score{}, err
			}
			s.G.Reverse(p)
			if score.Diff() < bestDiff {
				bestMove = move
				bestScore = score
				bestDiff = score.Diff()
			}
			if bestDiff < beta {
				beta = bestDiff
			}
			if beta <= alpha {
				break
			}
		}
		return bestMove, bestScore, nil

	default:
		panic("invalid player")
	}
}
