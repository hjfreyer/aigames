package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"
)

type TestInputs struct {
	Fields []string
}

type TestOutputs struct {
	Outputs []*Output
}

type Output struct {
	Elapsed time.Duration
	Move    string
	Score   Score
}

func ReadJson(path string, val interface{}) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	dec := json.NewDecoder(f)
	if err := dec.Decode(&val); err != nil {
		return err
	}
	return nil
}

func RunTests(inputs *TestInputs) (*TestOutputs, error) {
	res := &TestOutputs{
		Outputs: make([]*Output, len(inputs.Fields)),
	}

	for idx, in := range inputs.Fields {
		o, err := runTest(in)
		if err != nil {
			return nil, err
		}
		res.Outputs[idx] = o
	}

	return res, nil
}

func runTest(field string) (*Output, error) {
	var g Connect4Game
	s := Searcher{G: &g}

	g.ResetGame(field)

	const testDepth = 10
	start := time.Now()
	for i := 0; i < testDepth; i++ {
		if err := s.SearchToDepth(i, nil); err != nil {
			return nil, err
		}
	}
	elapsed := time.Since(start)

	move := "undefined"
	if s.Round.BestMove != nil {
		move = s.Round.BestMove.Encode()
	}

	log.Print(s)
	return &Output{
		Elapsed: elapsed,
		Move:    move,
		Score:   s.Round.BestScore,
	}, nil
}

func DiffOuts(actual, golden *TestOutputs) error {
	count := len(actual.Outputs)
	if len(golden.Outputs) < count {
		count = len(golden.Outputs)
	}

	for idx := 0; idx < count; idx++ {
		if err := diffOut(actual.Outputs[idx], golden.Outputs[idx]); err != nil {
			return err
		}
	}

	if len(actual.Outputs) != len(golden.Outputs) {
		return fmt.Errorf("Output matches golden, except actual has %d results, golden had %d",
			len(actual.Outputs), len(golden.Outputs))
	}
	return nil
}

func diffOut(actual, golden *Output) error {
	timeDiff := actual.Elapsed - golden.Elapsed
	if timeDiff < 0 {
		timeDiff = -timeDiff
	}
	if 0.1 < float64(timeDiff)/float64(golden.Elapsed) {
		return fmt.Errorf("Elapsed time more than 10%% off. Actual: %v; Golden %v",
			actual.Elapsed, golden.Elapsed)
	}
	return nil
}
