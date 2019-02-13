package runner

import (
	"cloud.google.com/go/pubsub"
	"context"
	"io"
	"log"
	"playback/config"
	"playback/file"
	"playback/util"
	"sync"
	"time"
)

// PlayRelative sets the window boundary to a lookahead duration value, reads the
// data from the input file line by line into memory and spawns the given action on
// the input data. The process is repeated until the EOF is met, or until the first
// timestamp outside the boundary is found. The thread then waits until the runtime
// clock is also outside the boundary (adjusted for an arbitrary jitter), shifts
// the boundary forward by the lookahead duration value and repeats.
// This method blocks until all lines and all spawned actions are completed.
//
// The parameters are the input reader implementation, action function, lookahead
// duration value and a maximum jitter setting (in milliseconds).
func PlayRelative(in file.Reader, action func(time.Time, []byte), lh time.Duration, mjMSec int) {
	delta := time.Duration(0)
	boundary := time.Now().Add(lh)

	log.Printf("Lookahead duration is %q with max jitter %q", lh, util.MSecToDuration(mjMSec))

	var wg sync.WaitGroup

	for {
		ts, d, e := in.ReadLine()
		if e == io.EOF {
			break
		}
		if e != nil {
			util.Fatal(e)
			return
		}
		if delta == 0 {
			delta = time.Now().Sub(ts)
			log.Printf("First timestamp is %q (delta vs now is %q)", ts, delta)
		}
		adjustedTS := ts.Add(delta)

		for adjustedTS.After(boundary) {
			jitter := util.Jitter(mjMSec)
			// wait until we're outside the window boundary + jitter
			<-time.After(boundary.Add(jitter).Sub(time.Now()))
			boundary = time.Now().Add(lh)
		}

		wg.Add(1)
		go func() {
			action(ts, d)
			wg.Done()
		}()
	}
	wg.Wait()
}

// PlayPaced reads the data from the input file line by line into memory
// and spawns the given action on the input data at a given rate until the EOF
// is met. The pacing is achieved by waiting the given delay duration
// between reads.
// This method blocks until all lines and all spawned actions are completed.
//
// The parameters are the input reader implementation, action function, delay
// duration value and a maximum jitter setting (in milliseconds).
func PlayPaced(in file.Reader, action func(time.Time, []byte), del time.Duration, mjMSec int) {
	var wg sync.WaitGroup

	log.Printf("Base delay between messages is %q with max jitter %q",
		del, util.MSecToDuration(mjMSec))

	for {
		ts, d, e := in.ReadLine()
		if e == io.EOF {
			break
		}
		if e != nil {
			util.Fatal(e)
			return
		}

		wg.Add(1)
		go func() {
			action(ts, d)
			wg.Done()
		}()

		jitter := util.Jitter(mjMSec)
		time.Sleep(time.Duration(jitter.Nanoseconds() + del.Nanoseconds()))
	}
	wg.Wait()
}

// PlayInstant attempts to read all the data from the input file line by
// line and spawn the given action on the input data. No throttling of limiting
// is implemented, hence the performance of this method is limited by the IO
// constraints, allocated memory and available lCPU.
// This method blocks until all lines and all spawned actions are completed.
//
// The parameters are the input reader implementation and action function.
func PlayInstant(in file.Reader, action func(time.Time, []byte)) {
	var wg sync.WaitGroup

	for {
		ts, d, e := in.ReadLine()
		if e == io.EOF {
			break
		}
		if e != nil {
			util.Fatal(e)
			return
		}

		wg.Add(1)
		go func() {
			action(ts, d)
			wg.Done()
		}()
	}
	wg.Wait()
}

// Output returns preconfigured publishing action function.
func Output(t *pubsub.Topic, c *config.ProgramConfig) func(time.Time, []byte) {
	return func(ts time.Time, d []byte) {
		publish(t, ts, d, c.Timeout)
	}
}

// publish handles PubSub publishing procedure with logging and allowing errors.
func publish(t *pubsub.Topic, ts time.Time, d []byte, to time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), to)
	defer cancel()

	res := t.Publish(ctx, &pubsub.Message{Data: d})
	id, e := res.Get(ctx)
	if e != nil {
		log.Printf("Error publishing message: %s", e)
		return
	}
	log.Printf("Published message %q (absolute timestamp %q)", id, ts)
}
