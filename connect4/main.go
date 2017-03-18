package main

import (
	// read from engine

	"fmt"
	"log"
	"math"
	_ "math/rand" // play randomly
	"os"          // read from engine
	_ "strconv"   // convert string to int
	_ "strings"   // split and replace function for strings
	"sync"
	"time" // for rand.seed(stuff)
)

var (
	Inf = math.Inf(1)
)

const (
	bufferTime  = time.Millisecond * 50
	timePerMove = time.Millisecond * 500
)

type gameState struct {
	Field        []byte
	PlayerToMove byte
	Heights      []int
	WinConds     []WinCondition
}

type WinCondition struct {
	Player1Count int
	Player2Count int
}

func NewGame(field []byte, playerToMove byte) *gameState {
	g := &gameState{
		Field:        field,
		PlayerToMove: playerToMove,
		Heights:      make([]int, 7),
		WinConds:     make([]WinCondition, 69),
	}

	for i, p := range field {
		if p != '0' {
			g.put(i, p, 1)
			g.Heights[i%7]++
		}
	}
	return g
}

//
// func (g *gameState) Clone() *gameState {
// 	n := &gameState{
// 		Field:        make([]byte, 42),
// 		PlayerToMove: g.PlayerToMove,
// 		Heights:      make([]int, 7),
// 		Quads:        make([]int, 69),
// 	}
// 	copy(n.Field, g.Field)
// 	copy(n.Heights, g.Heights)
// 	copy(n.Quads, g.Quads)
// 	return n
// }

// func sameSign(a, b int) bool {
// 	return (a < 0 && b < 0) || (0 < a && 0 < b)
// }

type Patch struct {
	LastPos int
}

func (g *gameState) put(pos int, player byte, delta int) {
	for _, quad := range spaceToQuads[pos] {
		if player == '1' {
			g.WinConds[quad].Player1Count += delta
		} else {
			g.WinConds[quad].Player2Count += delta
		}
	}
}

func (g *gameState) Move(col int) Patch {
	pos := col + 7*(5-g.Heights[col])
	if pos < 0 {
		panic("bad move")
	}
	g.Field[pos] = g.PlayerToMove
	g.put(pos, g.PlayerToMove, 1)
	g.Heights[col]++
	g.PlayerToMove = opponent(g.PlayerToMove)

	return Patch{LastPos: pos}
}

func (g *gameState) Reverse(p Patch) {
	col := p.LastPos % 7

	g.PlayerToMove = opponent(g.PlayerToMove)
	g.Field[p.LastPos] = '0'
	g.put(p.LastPos, g.PlayerToMove, -1)
	g.Heights[col]--
}

func evalNode(g *gameState) float64 {
	var p1score, p2score float64
	for _, wc := range g.WinConds {
		if wc.Player1Count > 0 && wc.Player2Count > 0 {
			continue
		}
		if wc.Player1Count == 4 {
			return -Inf
		}
		if wc.Player2Count == 4 {
			return Inf
		}
		if wc.Player1Count > 0 {
			p1score += math.Pow(10, float64(wc.Player1Count-1))
		}
		if wc.Player2Count > 0 {
			p2score += math.Pow(10, float64(wc.Player2Count-1))
		}
	}

	// Draw.
	if g.Full() {
		return 0
	}

	return p2score - p1score
}

func opponent(p byte) byte {
	switch p {
	case '1':
		return '2'
	case '2':
		return '1'
	default:
		panic("Weird id! ")
	}
}

func parseField(s string) []byte {
	res := make([]byte, 42)
	for i := range res {
		res[i] = s[2*i]
	}
	return res
}

func (g *gameState) Full() bool {
	for _, h := range g.Heights {
		if h < 6 {
			return false
		}
	}
	return true
}

func minimax(g *gameState, depth int, alpha, beta float64, t *ApproximateTimer) (Move, bool) {
	if t.DeadlineExceeded() {
		return Move{}, true
	}

	if depth == 0 {
		return Move{-1, evalNode(g)}, false
	}

	if e := evalNode(g); e == +Inf || e == -Inf {
		return Move{-1, e}, false
	}

	if g.Full() {
		return Move{-1, 0}, false
	}

	if g.PlayerToMove == '2' {
		bestMove := Move{-1, -Inf}
		for col := 0; col < 7; col++ {
			if g.Heights[col] == 6 {
				continue
			}
			p := g.Move(col)
			move, cancelled := minimax(g, depth-1, alpha, beta, t)
			if cancelled {
				return Move{}, true
			}
			g.Reverse(p)
			if bestMove.Val < move.Val {
				bestMove = move
				bestMove.Col = col
			}
			if alpha < move.Val {
				alpha = move.Val
			}
			if beta <= alpha {
				break
			}
		}
		return bestMove, false
	} else {
		bestMove := Move{-1, +Inf}
		for col := 0; col < 7; col++ {
			if g.Heights[col] == 6 {
				continue
			}
			p := g.Move(col)
			move, cancelled := minimax(g, depth-1, alpha, beta, t)
			if cancelled {
				return Move{}, true
			}
			g.Reverse(p)
			if move.Val < bestMove.Val {
				bestMove = move
				bestMove.Col = col
			}
			if move.Val < beta {
				beta = move.Val
			}
			if beta <= alpha {
				break
			}
		}
		return bestMove, false
	}
}

func getDeadline(timeBank time.Duration, roundNum, maxRounds int) time.Duration {
	bonusTime := timeBank - timePerMove
	res := timePerMove + (bonusTime / time.Duration(maxRounds-roundNum))
	if timeBank-bufferTime < res {
		res = timeBank - bufferTime
	}
	return res
}

func main() {
	msgs := make(chan interface{}, 10)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		if err := ParseAll(os.Stdin, NewConnect4Parser(), msgs); err != nil {
			log.Fatal(err)
		}
		wg.Done()
	}()

	var round int
	var field []byte
	var env GameEnv
	for msg := range msgs {
		switch m := msg.(type) {
		case *SettingsChange:
			if m.Attr == "your_botid" {
				env.PlayerID = m.Val[0]
			}

		case *RoundUpdate:
			round = m.Round
			log.Printf("BEGIN ROUND %d", round)

		case *FieldUpdate:
			field = m.NewField

		case *MoveRequest:
			env.Reset(field)
			moveDuration := getDeadline(m.TimeBank, round, 43)
			dl := time.Now().Add(moveDuration)
			log.Printf("Time Bank: %v, allocating %v", m.TimeBank, moveDuration)
			move, depth := env.NextMove(dl)
			log.Printf("Took %v more than the deadline", time.Now().Sub(dl))
			if (move.Val == -Inf && env.PlayerID == '1') || (move.Val == Inf && env.PlayerID == '2') {
				env.AlreadyWon = true
				env.AlreadyWonDepth = depth
			}
			log.Printf("Value: %v", move.Val)
			log.Printf("END OF ROUND %d", round)
			fmt.Printf("place_disc %d\n", move.Col)
		}
	}

	wg.Wait()
}
