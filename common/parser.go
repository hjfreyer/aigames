package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
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
