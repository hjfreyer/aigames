package main

import (
	"errors"
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

type Move interface {
	Encode() string
}
type Patch interface{}
type Cancelable interface {
	DeadlineExceeded() bool
}

type GameEngine interface {
	ResetGame(state string)

	GuessScore() Score
	CurrentPlayer() Player

	ValidMoves() []Move
	Move(Move) Patch
	Reverse(Patch)
}

type Searcher struct {
	G     GameEngine
	Round Round
}

type Round struct {
	BestMove  Move
	BestScore Score
	BestDepth int

	LastScore Score
	LastDepth int
}

func (s *Searcher) Reset(state string) {
	s.G.ResetGame(state)
	s.Round = Round{}

	// Start with the worst possible score.
	s.Round.BestScore.GameOver = true
	switch s.G.CurrentPlayer() {
	case Player1:
		s.Round.BestScore.Player2 = math.Inf(1)
	case Player2:
		s.Round.BestScore.Player1 = math.Inf(1)
	default:
		panic("bad player")
	}
	s.Round.LastScore = s.Round.BestScore
}

func (e *Searcher) SearchToDepth(depth int, c Cancelable) error {
	move, score, err := e.minimax(depth, math.Inf(-1), math.Inf(1), c)
	if err != nil {
		return err
	}
	e.Round.LastScore = score
	e.Round.LastDepth = depth
	if !score.IsWinForPlayer(Opponent(e.G.CurrentPlayer())) {
		e.Round.BestMove = move
		e.Round.BestScore = score
		e.Round.BestDepth = depth
	}
	return nil
}

func (e *Searcher) NextMove(deadline time.Time) error {
	depth := 1
	var t ApproximateTimer
	t.Start(deadline)
	for {
		if err := e.SearchToDepth(depth, &t); err == ErrDeadlineExceeded {
			return nil
		} else if err != nil {
			return err
		}

		if e.isWin(e.Round.LastScore) {
			break
		}

		depth++
	}

	return nil
}

func (s *Searcher) isWin(score Score) bool {
	switch s.G.CurrentPlayer() {
	case Player1:
		return math.IsInf(score.Player1, 1)
	case Player2:
		return math.IsInf(score.Player2, 1)
	default:
		panic("Bad player value")
	}
}

func (s *Searcher) minimax(depth int, alpha, beta float64, t Cancelable) (Move, Score, error) {
	if t != nil && t.DeadlineExceeded() {
		return nil, Score{}, ErrDeadlineExceeded
	}

	score := s.G.GuessScore()
	if depth == 0 || score.GameOver {
		return nil, score, nil
	}

	switch s.G.CurrentPlayer() {
	case Player2:
		var bestMove Move
		bestScore := Score{GameOver: true, Player1: math.Inf(1), Player2: 0}

		for _, move := range s.G.ValidMoves() {
			p := s.G.Move(move)
			_, score, err := s.minimax(depth-1, alpha, beta, t)
			s.G.Reverse(p)
			if err != nil {
				return nil, Score{}, err
			}
			if bestScore.Diff() < score.Diff() {
				bestMove = move
				bestScore = score
			}
			if alpha < bestScore.Diff() {
				alpha = bestScore.Diff()
			}
			if beta <= alpha {
				break
			}
		}
		return bestMove, bestScore, nil
	case Player1:
		var bestMove Move
		bestScore := Score{GameOver: true, Player1: 0, Player2: math.Inf(1)}

		for _, move := range s.G.ValidMoves() {
			p := s.G.Move(move)
			_, score, err := s.minimax(depth-1, alpha, beta, t)
			s.G.Reverse(p)
			if err != nil {
				return nil, Score{}, err
			}
			if score.Diff() < bestScore.Diff() {
				bestMove = move
				bestScore = score
			}
			if bestScore.Diff() < beta {
				beta = bestScore.Diff()
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
