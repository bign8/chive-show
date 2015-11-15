package chain

import "sync"

// Worker is a function designed to fan out and perform work on a piece of Data
type Worker func(in <-chan interface{}, out chan<- interface{}, idx int)

// FanOut allows lengthy workers to fan out on chanel operations
func FanOut(count int, buff int, in <-chan interface{}, doIt Worker) <-chan interface{} {
	out := make(chan interface{}, buff)
	var wg sync.WaitGroup
	wg.Add(count)
	for i := 0; i < count; i++ {
		go func(idx int) {
			doIt(in, out, idx)
			wg.Done()
		}(i)
	}
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

// FanIn takes multiple chanels and pushes their results into a single channel
func FanIn(buff int, cs ...<-chan interface{}) <-chan interface{} {
	var wg sync.WaitGroup
	out := make(chan interface{})

	// Start an output goroutine for each input channel in cs.  output
	// copies values from c to out until c is closed, then calls wg.Done.
	output := func(c <-chan interface{}) {
		for n := range c {
			out <- n
		}
		wg.Done()
	}
	wg.Add(len(cs))
	for _, c := range cs {
		go output(c)
	}

	// Start a goroutine to close out once all the output goroutines are
	// done.  This must start after the wg.Add call.
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}
