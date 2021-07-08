package models

// Is this Data?

func NewRecord(name string, data []byte) Record {
  return Record{
    Type: name,
    Data: data,
  }
}

type Record struct {
  Type string
  Data []byte
}
