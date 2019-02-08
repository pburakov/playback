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
		}
		if delta == 0 {
			delta = time.Now().Sub(ts)
			log.Printf("First timestamp is %q (delta vs now is %q)", ts, delta)
		}
		adjustedTS := ts.Add(delta)

		for adjustedTS.After(boundary) {
			jitter := util.Jitter(mjMSec)
			<-time.After(boundary.Add(jitter).Sub(time.Now())) // wait until we're outside the window boundary + jitter
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

func PlayPaced(in file.Reader, action func(time.Time, []byte), del time.Duration, mjMSec int) {
	var wg sync.WaitGroup

	log.Printf("Base delay between messages is %q with max jitter %q", del, util.MSecToDuration(mjMSec))

	for {
		ts, d, e := in.ReadLine()
		if e == io.EOF {
			break
		}
		if e != nil {
			util.Fatal(e)
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

func PlayInstant(in file.Reader, action func(time.Time, []byte)) {
	var wg sync.WaitGroup

	for {
		ts, d, e := in.ReadLine()
		if e == io.EOF {
			break
		}
		if e != nil {
			util.Fatal(e)
		}

		wg.Add(1)
		go func() {
			action(ts, d)
			wg.Done()
		}()
	}
	wg.Wait()
}

func Output(t *pubsub.Topic, c *config.ProgramConfig) func(time.Time, []byte) {
	return func(ts time.Time, d []byte) {
		publish(t, ts, d, c.Timeout)
	}
}

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
