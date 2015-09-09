package pipeline

import (
  "app/helpers/pipeline/models"
  "net/http"
)

// https://www.adayinthelifeof.nl/2011/06/02/asynchronous-operations-in-rest/
// http://restcookbook.com/Resources/asynchroneous-operations/

func (pl *PipelineMaster) httpBind(base string) {
  http.HandleFunc(base + "/start", pl.httpInit)
  http.HandleFunc(base + "/queue/", pl.httpProgress)
  http.HandleFunc(base + "/result/", pl.httpResult)
}

func (pl *PipelineMaster) httpInit(w http.ResponseWriter, r *http.Request) {
  // TODO: process payload
  pl.Inject(models.StreamRecord{})

  // REDIRECT 202
  // Location: base + "/queue/" + jobID
}

func (pl *PipelineMaster) httpProgress(w http.ResponseWriter, r *http.Request) {
  // TODO: process progress request

  // IF no trailing jobID
  //   return list of all currently running processes (admin)
  // ELIF trailing jobID DNE
  //   return 404
  // ELIF trailing jobID is Processing
  //   return 200 for status
  // ELIF trailing jobID is done
  //   REDIRECT 303
  //   Location: base + "/result/" + resultID
}

func (pl *PipelineMaster) httpResult(w http.ResponseWriter, r *http.Request) {
  // TODO: process result request
  // 200 JSON response object
}
