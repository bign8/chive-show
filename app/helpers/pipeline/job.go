package main

import (
  "./models"
  "math/rand"
)

type JobStatus uint
type JobID uint32

const (
  PENDING  JobStatus = 0
  COMPLETE JobStatus = 1
  FAILURE  JobStatus = 2
)

func generateJobID() JobID {
  return JobID(rand.Uint32())
}

func NewJob(id JobID) *Job {
  return &Job{
    Id: id,
    Status: PENDING,
    Progress: 0,
    Result: nil,
  }
}

// TODO: add serialize helpers here
type Job struct {
  Id       JobID
  Status   JobStatus
  Progress uint
  Result   []byte
  watchers []func(string, *Job)
  Type     string
}

func (j *Job) updateProgress(record models.StreamRecord) {
  start_progress := j.Progress

  // TODO: compute overall job process based on type and previous emitted StreamRecords

  if j.Progress == 100 && j.Result != nil {
    j.dispatch("complete")
  } else if j.Progress != start_progress {
    j.dispatch("update")
  }
}

func (j *Job) dispatch(event string) {
  // TODO: dispatch every 100 ms
  for _, watcher := range j.watchers {
    go watcher(event, j)
  }
}

func (j *Job) Watch(watcher func(string, *Job)) {
  if j.watchers == nil {
    j.watchers = make([]func(string, *Job), 0)
    // TODO: start job watcher daemon (for dispatch every 100 ms if necessary)
  }
  j.watchers = append(j.watchers, watcher)
}
