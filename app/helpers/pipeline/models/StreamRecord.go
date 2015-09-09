package models

func NewStreamRecord(stream string, data Record) *StreamRecord {
  return &StreamRecord{
    Stream: stream,
    Record: data,
  }
}

type StreamRecord struct {
  Stream string
  Record Record
}
