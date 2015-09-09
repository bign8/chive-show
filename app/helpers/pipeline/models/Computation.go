package models

import "errors"

var NotImplementedError = errors.New("Not Implemented")

type IComputation interface {
  ProcessRecord(data Record) error
  ProcessTimer(timer Timer) error
}

type Computation struct {
  IComputation
  State *State
  out chan<- StreamRecord
}

func (c *Computation) ProcessRecord(data Record) error {
  return NotImplementedError
}

func (c *Computation) ProcessTimer(timer Timer) error {
  return NotImplementedError
}

func (c *Computation) SetTimer(timer Timer) {
  // TODO: this can consume a ton of memory... if called tons of times
  go func() {
    <- timer.channel
    c.ProcessTimer(timer)
  }()
}

func (c *Computation) ProduceRecord(data Record, stream string) {
  c.out <- StreamRecord{
    Record: data,
    Stream: stream,
  }
}

func (c *Computation) run(in <-chan Record, out chan<- StreamRecord) (err error) {
  c.out = out
  defer close(c.out)
  for record := range in {
    err = c.ProcessRecord(record)
    if err != nil {
      return
    }
  }
  return nil
}
