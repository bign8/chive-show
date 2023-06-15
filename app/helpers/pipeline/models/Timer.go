package models

import "time"

func NewTimer(d time.Duration, data Record, recuring bool) Timer {
  timer := Timer{Data: data}
  if recuring {
    timer.channel = time.Tick(d)
  } else {
    timer.channel = time.After(d)
  }
  return timer
}

type Timer struct {
  Data Record
  channel <-chan time.Time
}
