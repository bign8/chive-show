package JobManager

import (
  "math/rand"
  "sync"
  "errors"
)

var (
  // jobs = make(map[JobID]*Job)
  // lock = sync.RWMutex
  ErrJobNotFound = errors.New("Job Not Found")
)

type _manager struct {
  sync.RWMutex
  jobs map[JobID]*Job
}

var manager = _manager{
  jobs: make(map[JobID]*Job),
}

func generateJobID() JobID {
  id := JobID(rand.Uint32())
  manager.RLock()
  _, ok := manager.jobs[id]
  for ok {
    id := JobID(rand.Uint32())
    _, ok = manager.jobs[id]
  }
  manager.RUnlock()
  return id
}

func CreateJob() *Job {
  id := generateJobID()
  new_job := &Job{
    Id: id,
    Status: PENDING,
    Progress: 0,
    Result: nil,
  }
  manager.Lock()
  manager.jobs[id] = new_job
  manager.Unlock()
  return new_job
}

func DeleteJob(id JobID) error {
  manager.RLock()
  job, ok := manager.jobs[id]
  manager.RUnlock()
  if !ok {
    return ErrJobNotFound
  }
  job.dispatch("delete")
  manager.Lock()
  delete(manager.jobs, id)
  manager.Unlock()
  return nil
}

func GetJob(id JobID) (*Job, error) {
  manager.RLock()
  job, ok := manager.jobs[id]
  manager.RUnlock()
  if !ok {
    return nil, ErrJobNotFound
  }
  return job, nil
}
