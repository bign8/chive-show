package crawler

// Batcher takes input and batches to given sizes
func Batcher(in <-chan string, size int) <-chan []string {
	out := make(chan []string)
	go func() {
		defer close(out)
		batch := make([]string, size)
		count := 0
		for post := range in {
			batch[count] = post
			count++
			if count >= size {
				count = 0
				out <- batch
				batch = make([]string, size) // allocate another chunk of memory
			}
		}
		if count > 0 {
			out <- batch[:count]
		}
	}()
	return out
}
