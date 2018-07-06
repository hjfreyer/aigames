package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"time"
)

const (
	bufferTime  = time.Millisecond * 50
	timePerMove = time.Millisecond * 500
)

// Specific stuff.

type SettingsChange struct {
	Attr string
	Val  string
}

func parseSettingsChange(s *bufio.Scanner) (interface{}, error) {
	var res SettingsChange
	if !s.Scan() {
		return nil, errors.New("No attribute found")
	}
	res.Attr = s.Text()
	if !s.Scan() {
		return nil, errors.New("No value found")
	}
	res.Val = s.Text()
	return &res, nil
}

type RoundUpdate struct {
	Round int
}

func parseRoundUpdate(s *bufio.Scanner) (interface{}, error) {
	if !s.Scan() {
		return nil, UnexpectedEnd
	}
	r, err := strconv.Atoi(s.Text())
	if err != nil {
		return nil, err
	}
	return &RoundUpdate{Round: r}, nil
}

type FieldUpdate struct {
	NewField string
}

func parseFieldUpdate(s *bufio.Scanner) (interface{}, error) {
	if !s.Scan() {
		return nil, UnexpectedEnd
	}
	return &FieldUpdate{NewField: s.Text()}, nil
}

type MoveRequest struct {
	TimeBank time.Duration
}

func parseMoveRequest(s *bufio.Scanner) (interface{}, error) {
	if !s.Scan() {
		return nil, UnexpectedEnd
	}
	n, err := strconv.Atoi(s.Text())
	if err != nil {
		return nil, err
	}
	return &MoveRequest{TimeBank: time.Millisecond * time.Duration(n)}, nil
}

func NewConnect4Parser() MessageParser {
	return CommandGroupParser{
		"settings": ParserFunc(parseSettingsChange),
		"update": CommandGroupParser{
			"game": CommandGroupParser{
				"round": ParserFunc(parseRoundUpdate),
				"field": ParserFunc(parseFieldUpdate),
			},
		},
		"action": CommandGroupParser{
			"move": ParserFunc(parseMoveRequest),
		},
	}
}

func getDeadline(timeBank time.Duration, roundNum, maxRounds int) time.Duration {
	bonusTime := timeBank - timePerMove
	bonusTime /= time.Duration(maxRounds - roundNum)
	if roundNum < 20 {
		bonusTime *= 3
	}
	res := timePerMove + bonusTime
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

	var ge Connect4Game
	searcher := Searcher{G: &ge}

	var round int
	var field string
	for msg := range msgs {
		switch m := msg.(type) {
		case *SettingsChange:
			if m.Attr == "your_botid" {
				if m.Val == "1" {
					log.Print("Player: 1. -Inf is our goal.")
				} else {
					log.Print("Player: 2. +Inf is our goal.")
				}
			}

		case *RoundUpdate:
			round = m.Round
			log.Printf("BEGIN ROUND %d", round)

		case *FieldUpdate:
			field = m.NewField

		case *MoveRequest:
			searcher.Reset(field)
			moveDuration := getDeadline(m.TimeBank, round, 43)
			dl := time.Now().Add(moveDuration)
			log.Printf("Time Bank: %v, allocating %v", m.TimeBank, moveDuration)
			err := searcher.NextMove(dl)
			if err != nil {
				log.Fatal(err)
			}
			log.Printf("Value: %v", searcher.Round.BestScore.Diff())
			fmt.Printf("place_disc %d\n", searcher.Round.BestMove.(Connect4Move).Column)
		}
	}

	wg.Wait()
}
