package main

import (
	"log"
	"math"
	"strconv"
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

type WinCondition struct {
	Player1Count int
	Player2Count int
}

type Connect4Game struct {
	Heights    []int
	WinConds   []WinCondition
	TotalMoves int
}

func (g *Connect4Game) ResetGame(state string) {
	g.Heights = make([]int, 7)
	g.WinConds = make([]WinCondition, 69)
	g.TotalMoves = 0
	for i := 0; i < 42; i++ {
		var p Player
		switch state[i*2] {
		case '0':
			continue
		case '1':
			p = Player1
		case '2':
			p = Player2
		default:
			panic("Bad field")
		}
		g.delta(i, p, 1)
	}
}

func (g *Connect4Game) delta(pos int, player Player, amt int) {
	g.TotalMoves += amt
	g.Heights[pos%7] += amt
	for _, quad := range spaceToQuads[pos] {
		switch player {
		case Player1:
			g.WinConds[quad].Player1Count += amt
		case Player2:
			g.WinConds[quad].Player2Count += amt
		default:
			panic("Bad Player")
		}
	}
}

func (g *Connect4Game) GuessScore() Score {
	var p1score, p2score float64
	for _, wc := range g.WinConds {
		if wc.Player1Count > 0 && wc.Player2Count > 0 {
			continue
		}
		if wc.Player1Count == 4 {
			return Score{GameOver: true, Player1: math.Inf(1), Player2: 0}
		}
		if wc.Player2Count == 4 {
			return Score{GameOver: true, Player1: 0, Player2: math.Inf(1)}
		}
		if wc.Player1Count > 0 {
			p1score += math.Pow(10, float64(wc.Player1Count-1))
		}
		if wc.Player2Count > 0 {
			p2score += math.Pow(10, float64(wc.Player2Count-1))
		}
	}

	// Draw.
	if g.TotalMoves == 42 {
		return Score{GameOver: true}
	}

	return Score{GameOver: false, Player1: p1score, Player2: p2score}
}

func (g *Connect4Game) CurrentPlayer() Player {
	if g.TotalMoves%2 == 0 {
		return Player1
	} else {
		return Player2
	}
}

type Connect4Move struct{ Column int }

func (g Connect4Move) Encode() string {
	return strconv.Itoa(g.Column)
}

type Connect4Patch struct{ position int }

func (g *Connect4Game) ValidMoves() []Move {
	var res []Move
	for i := 0; i < 7; i++ {
		if g.Heights[i] < 6 {
			res = append(res, Connect4Move{i})
		}
	}
	return res
}

func (g *Connect4Game) Move(m Move) Patch {
	col := m.(Connect4Move).Column
	pos := col + 7*(5-g.Heights[col])
	if pos < 0 {
		panic("bad move")
	}
	g.delta(pos, g.CurrentPlayer(), 1)
	return Connect4Patch{pos}
}

func (g *Connect4Game) Reverse(p Patch) {
	pos := p.(Connect4Patch).position
	g.delta(pos, Opponent(g.CurrentPlayer()), -1)
}
