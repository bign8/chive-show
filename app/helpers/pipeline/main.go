package main

import (
  // "flag"
  "./JobManager"
  "encoding/json"
  "fmt"
  // "html"
  "log"
  "net/http"
  // "os"
  // "os/signal"
  // "syscall"
)

func main() {
  // sigs := make(chan os.Signal, 1)
  // done := make(chan bool, 1)
  //
  // signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
  // go func() {
  //   for {
  //     sig := <-sigs
  //     fmt.Println()
  //     fmt.Println(sig)
  //     done <- true
  //   }
  // }()
  //
  // fmt.Println("awaiting signal")
  // <-done
  // fmt.Println("exiting")
  // <-done
  // fmt.Println("really leaving")
  // <-done
  // fmt.Println("really leaving")


  // master := NewPipelineMaster()
  // done = make(chan struct{}, 1)
  // go master.Main(done)

  http.HandleFunc("/queue/", func(w http.ResponseWriter, r *http.Request) {
    // fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))

    w.Header().Set("Content-Type", "application/json; charset=utf-8")

    job := JobManager.Job{
      Id:       99,
      Status:   JobManager.PENDING,
      Progress: 94,
      Result:   []byte{1,2,3},
    }
    data, err := json.MarshalIndent(job, "", "  ")
    if err != nil {
      fmt.Fprintf(w, "error %v", err)
    } else {
      fmt.Fprint(w, string(data))
    }
  })
  log.Fatal(http.ListenAndServe(":8081", nil))
}
