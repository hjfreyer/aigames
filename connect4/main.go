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
	_ "time" // for rand.seed(stuff)
)

var spaceToQuads [][]int

var (
	Inf = math.Inf(1)
)

func buildReference() [][]int {
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
	return res
}

type gameState struct {
	Field        []byte
	PlayerToMove byte
	Heights      []int
	Quads        []int
}

func NewGame(field []byte, playerToMove byte) *gameState {
	g := &gameState{
		Field:        field,
		PlayerToMove: playerToMove,
		Heights:      make([]int, 7),
		Quads:        make([]int, 69),
	}

	for i, p := range field {
		if p != '0' {
			g.put(i, p)
			g.Heights[i%7]++
		}
	}
	return g
}

func (g *gameState) Clone() *gameState {
	n := &gameState{
		Field:        make([]byte, 42),
		PlayerToMove: g.PlayerToMove,
		Heights:      make([]int, 7),
		Quads:        make([]int, 69),
	}
	copy(n.Field, g.Field)
	copy(n.Heights, g.Heights)
	copy(n.Quads, g.Quads)
	return n
}

func sameSign(a, b int) bool {
	return (a < 0 && b < 0) || (0 < a && 0 < b)
}

func (g *gameState) put(pos int, player byte) {
	var sign int
	if player == '1' {
		sign = -1
	} else {
		sign = 1
	}

	for _, quad := range spaceToQuads[pos] {
		if g.Quads[quad] == 5 {
			// Change nothing
		} else if g.Quads[quad] == 0 {
			// Start off the process
			g.Quads[quad] = sign
		} else if sameSign(g.Quads[quad], sign) {
			g.Quads[quad] += sign
		} else {
			g.Quads[quad] = 5
		}
	}
}

func (g *gameState) Move(col int) {
	pos := col + 7*(5-g.Heights[col])
	if pos < 0 {
		panic("bad move")
	}
	g.Field[pos] = g.PlayerToMove
	g.put(pos, g.PlayerToMove)
	g.Heights[col]++
	g.PlayerToMove = opponent(g.PlayerToMove)
}

func evalNode(g *gameState) float64 {
	var empty, p1total, p2total int

	p1counts := make([]int, 4)
	p2counts := make([]int, 4)
	for _, q := range g.Quads {
		switch {
		case q == 5:
			continue
		case q == 0:
			empty++
			continue
		case q == 4:
			return +Inf
		case q == -4:
			return -Inf
		case q < 0:
			p1total++
			p1counts[-q]++
		case 0 < q:
			p2total++
			p2counts[q]++
		}
	}

	// Draw.
	if p1total == 0 && p2total == 0 && empty == 0 {
		return 0
	}
	var p1score, p2score int
	for i := 3; 0 < i; i-- {
		p1score *= 10
		p2score *= 10
		p1score += p1counts[i]
		p2score += p2counts[i]
	}
	return float64(p2score - p1score)
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

//
//type scoreEstimate struct {
//	Depth int
//	Score byte
//}

//type thingy struct {
//	Cache map[string]scoreEstimate
//}

//func getWinner(g *gameState) byte {
//	for _, q := range g.Quads {/
//		switch q {
//		case -4:
//			return '1'
//		case 4:
//			return '2'
//		}
//	}
//	return '0'
//}

func (g *gameState) Full() bool {
	for _, h := range g.Heights {
		if h < 6 {
			return false
		}
	}
	return true
}

func minimax(g *gameState, depth int, alpha, beta float64) (int, float64) {
	if depth == 0 {
		return -1, evalNode(g)
	}

	if e := evalNode(g); e == +Inf || e == -Inf {
		return -1, e
	}

	if g.Full() {
		return -1, 0
	}

	if g.PlayerToMove == '2' {
		bestCol := -1
		bestValue := -Inf
		for col := 0; col < 7; col++ {
			if g.Heights[col] < 6 {
				nextG := g.Clone()
				nextG.Move(col)
				_, v := minimax(nextG, depth-1, alpha, beta)
				if bestValue < v {
					bestCol = col
					bestValue = v
				}
				if alpha < v {
					alpha = v
				}
				if beta <= alpha {
					break
				}
			}
		}
		return bestCol, bestValue
	} else {
		bestCol := -1
		bestValue := +Inf
		for col := 0; col < 7; col++ {
			if g.Heights[col] < 6 {
				nextG := g.Clone()
				nextG.Move(col)
				_, v := minimax(nextG, depth-1, alpha, beta)
				if v < bestValue {
					bestCol = col
					bestValue = v
				}
				if v < beta {
					beta = v
				}
				if beta <= alpha {
					break
				}
			}
		}
		return bestCol, bestValue
	}
}

func main() {
	spaceToQuads = buildReference()

	msgs := make(chan interface{}, 10)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		if err := ParseAll(os.Stdin, NewConnect4Parser(), msgs); err != nil {
			log.Fatal(err)
		}
		wg.Done()
	}()

	var id, round string
	var field []byte
	for msg := range msgs {
		switch m := msg.(type) {
		case *SettingsChange:
			if m.Attr == "your_botid" {
				id = m.Val
			}

		case *RoundUpdate:
			round = m.RoundID

		case *FieldUpdate:
			field = m.NewField

		case *MoveRequest:
			game := NewGame(field, id[0])
			c, v := minimax(game, 9, -Inf, Inf)
			log.Printf("Round: %s, Value: %v", round, v)
			fmt.Printf("place_disc %d\n", c)
		}
	}

	wg.Wait()
	return
	// 	if _, ok := msg.(*endOfFile); ok {
	// 		return
	// 	}
	// 	if ur, ok := msg.(*updateRound); ok {
	//
	// 		msg = readMessage(scan)
	// 	}
	//
	// 	field := msg.(*updateField).NewField
	// 	game := NewGame(field, id)
	// 	c, v := minimax(game, 9, -Inf, Inf)
	// 	log.Print("Value: ", v)
	//
	// 	_ = readMessage(scan).(*moveRequest)
	//
	// 	readMessage(scan) // New field
	// }
}

/*
settings timebank 10000
settings time_per_move 500
settings player_names player1,player2
settings your_bot player1
settings your_botid 1
settings field_columns 7
settings field_rows 6

update game round 1
update game field 0,0,0,0,0,0,0;0,0,0,0,0,0,0;0,0,0,0,0,0,0;0,0,0,0,0,0,0;0,0,0,0,0,0,0;0,0,0,0,0,0,0
action move 10000
*/
