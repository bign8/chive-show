package pipeline

import "app/helpers/pipeline/models"

func Run(configFile string) error {
  // TODO: do stuff
  // TODO: route traffic
  // TODO: spin up servers if necessary
  // TODO: spin up goroutines if necessary
  // TODO: all the things
  // TODO: key mining
  pl := NewPipelineMaster()

  // TODO: parse config

  go pl.Main()
  return nil
}

func NewPipelineMaster() *PipelineMaster {
  return &PipelineMaster{
    stream: make(chan models.StreamRecord),
    jobs:   make(map[JobID]*Job),
  }
}

type PipelineMaster struct {
  stream chan models.StreamRecord
  jobs   map[JobID]*Job
}

func (pl *PipelineMaster) Main() {
  // TODO: run pipelines
}

func (pl *PipelineMaster) Inject(data models.StreamRecord) {
  pl.stream <- data
}

func (pl *PipelineMaster) NewJob(task string, payload models.Record) JobID {
  // TODO: pre-allocate pipeline goroutines
  id := pl.newJobId()
  job := NewJob(id)
  pl.jobs[id] = job
  pl.Inject(*models.NewStreamRecord(task, payload))
  return job.Id
}

func (pl *PipelineMaster) newJobId() JobID {
  id := generateJobID()
  for pl.jobs[id] != nil {
    id = generateJobID()
  }
  return id
}

func (pl *PipelineMaster) GetJob(id JobID) *Job {
  return pl.jobs[id]
}

func (pl *PipelineMaster) DelJob(id JobID) {
  // TODO: figure out how to implement this
}
