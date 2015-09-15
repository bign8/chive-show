package main

import (
  "./models"
  "errors"
)

var (
  ErrStreamNotCreated = errors.New("Stream Not Created")
)

type StreamMaster struct {
  streams map[models.StreamTitle]chan models.Record
}

func NewStreamMaster() *StreamMaster {
  return &StreamMaster{
    streams: make(map[models.StreamTitle]chan models.Record),
  }
}

func (sr *StreamMaster) Route(stream_record models.StreamRecord) error {
  channel, ok := sr.streams[stream_record.Stream]
  if !ok {
    return ErrStreamNotCreated
  }
  channel <- stream_record.Record
  return nil
}

func (sr *StreamMaster) AddStream(title models.StreamTitle) (<-chan models.Record, error) {
  return nil, nil
}

type StreamRouter struct {
  miner   struct{} // TODO: document miner API
  streams map[string]*models.Computation
}
