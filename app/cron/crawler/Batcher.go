package crawler

// Batcher takes input and batches to given sizes
func Batcher(in <-chan Data, size int) <-chan []Data {
	out := make(chan []Data)
	go func() {
		defer close(out)
		batch := make([]Data, size)
		count := 0
		for post := range in {
			batch[count] = post
			count++
			if count >= size {
				count = 0
				out <- batch
				batch = make([]Data, size) // allocate another chunk of memory
			}
		}
		if count > 0 {
			out <- batch[:count]
		}
	}()
	return out
}
