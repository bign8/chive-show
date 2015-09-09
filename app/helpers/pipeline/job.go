package pipeline

import "math/rand"

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
}
