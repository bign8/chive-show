package crawler

func Batcher(in <-chan ChivePost, batch_size int) <-chan []ChivePost {
  out := make(chan []ChivePost)
  go func() {
    defer close(out)
    batch := make([]ChivePost, batch_size)
    count := 0
    for post := range in {
      batch[count] = post
      count++
      if count >= batch_size {
        count = 0
        out <- batch
        batch = make([]ChivePost, batch_size) // allocate another chunk of memory
      }
    }
    if count > 0 {
      out <- batch[:count]
    }
  }()
  return out
}
