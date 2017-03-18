package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"time"
)

// Generic Stuff.

var UnexpectedEnd = errors.New("Unexpected end of stream")

type MessageParser interface {
	Next(*bufio.Scanner) (interface{}, error)
}

type CommandGroupParser map[string]MessageParser

func (p CommandGroupParser) Next(s *bufio.Scanner) (interface{}, error) {
	if !s.Scan() {
		return nil, s.Err()
	}
	parser := p[s.Text()]
	if parser == nil {
		return nil, fmt.Errorf("Unknown command group %q", s.Text())
	}
	return parser.Next(s)
}

type ParserFunc func(*bufio.Scanner) (interface{}, error)

func (f ParserFunc) Next(s *bufio.Scanner) (interface{}, error) {
	return f(s)
}

func ParseAll(r io.Reader, p MessageParser, msgs chan<- interface{}) error {
	scan := bufio.NewScanner(r)
	scan.Split(bufio.ScanWords)

	defer close(msgs)

	for {
		val, err := p.Next(scan)
		if err != nil {
			return err
		}
		if val == nil {
			return nil
		}
		msgs <- val
	}
}

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
	NewField []byte
}

func parseFieldUpdate(s *bufio.Scanner) (interface{}, error) {
	if !s.Scan() {
		return nil, UnexpectedEnd
	}
	return &FieldUpdate{NewField: parseField(s.Text())}, nil
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
