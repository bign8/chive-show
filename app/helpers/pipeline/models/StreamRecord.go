package models

type StreamTitle string

func NewStreamTitle(stream interface{}) StreamTitle {
  return stream.(StreamTitle)
}

type StreamRecord struct {
  Stream StreamTitle
  Record Record
}

func NewStreamRecord(stream StreamTitle, data Record) *StreamRecord {
  return &StreamRecord{
    Stream: stream,
    Record: data,
  }
}
