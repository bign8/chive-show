package models

import "errors"

// TODO: properly serialize and de-serialize from store

func NewState() *State {
  return &State{make(map[string][]byte)}
}

type State struct {
  store map[string][]byte
}

func (s *State) Add(name string, data interface{}) error {
  bit_data, ok := data.([]byte)
  if ok {
    s.store[name] = bit_data
  } else {
    return errors.New("Typecaste to []byte invalid")
  }
  return nil
}

func (s *State) Get(name string) (interface{}, error) {
  data, ok := s.store[name]
  if !ok {
    return nil, errors.New("Default state")
  }
  return data, nil
}
