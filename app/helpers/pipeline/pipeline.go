package main

import (
  "./models"
  "./JobManager"
)

func Run(configFile string) error {
  // TODO: do stuff
  // TODO: route traffic
  // TODO: spin up servers if necessary
  // TODO: spin up goroutines if necessary
  // TODO: all the things
  // TODO: key mining
  // pl := NewPipelineMaster()

  // TODO: parse config

  // go pl.Main()
  return nil
}

func NewPipelineMaster() *PipelineMaster {
  return &PipelineMaster{
    stream: make(chan models.StreamRecord),
    jobs:   make(map[JobManager.JobID]*JobManager.Job),
  }
}

type PipelineMaster struct {
  stream chan models.StreamRecord
  jobs   map[JobManager.JobID]*JobManager.Job
}

func (pl *PipelineMaster) Main(done <-chan struct{}) {
  for {
    // run pipelines
  }
  // TODO: tear down
}

func (pl *PipelineMaster) Inject(data models.StreamRecord) {
  pl.stream <- data
}
