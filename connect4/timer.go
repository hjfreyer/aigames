package main

import (
	"time"
)

type ApproximateTimer struct {
	deadline  time.Time
	startTime time.Time
	ticks     int64

	nextCheck    int64
	maxIncrement int64
}

func (t *ApproximateTimer) Start(deadline time.Time) {
	t.deadline = deadline
	t.startTime = time.Now()
	t.ticks = 0

	t.nextCheck = 2
	t.maxIncrement = 2
}

func (t *ApproximateTimer) DeadlineExceeded() bool {
	//log.Printf("%+v", t)
	if t == nil {
		return false
	}

	t.ticks++
	if t.ticks < t.nextCheck {
		return false
	}
	now := time.Now()
	if now.After(t.deadline) {
		return true
	}
	timePerTick := now.Sub(t.startTime) / time.Duration(t.ticks)
	targetElapsed := t.deadline.Sub(t.startTime)
	ticksUntilTarget := int64(targetElapsed / timePerTick)
	//	log.Print("Ticks: ", t.ticks)
	//	log.Print("Estimated ticks: ", ticksUntilTarget)

	// Set the next checkpoint to halfway between now and the estimated target.
	t.nextCheck = ticksUntilTarget/2 + t.ticks/2

	if t.ticks+t.maxIncrement < t.nextCheck {
		t.nextCheck = t.ticks + t.maxIncrement
	}
	t.maxIncrement *= 2
	return false
}
