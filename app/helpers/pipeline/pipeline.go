package pipeline

import "app/helpers/pipeline/models"

func Run(configFile string) error {
  // TODO: do stuff
  // TODO: route traffic
  // TODO: spin up servers if necessary
  // TODO: spin up goroutines if necessary
  // TODO: all the things
  // TODO: key mining
  pl := NewPipeline()

  // TODO: parse config

  go pl.Main()
  return nil
}

func NewPipeline() *Pipeline {
  return &Pipeline{
    stream: make(chan models.StreamRecord),
  }
}

type Pipeline struct {
  stream chan models.StreamRecord
}

func (pl *Pipeline) Main() {
  // TODO: run pipelines
}

func (pl *Pipeline) Inject(data models.StreamRecord) {
  pl.stream <- data
}
