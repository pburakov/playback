package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/pburakov/playback/config"
	"github.com/pburakov/playback/input"
	"github.com/pburakov/playback/input/avro"
	"github.com/pburakov/playback/input/csv"
	"github.com/pburakov/playback/input/json"
	"github.com/pburakov/playback/output"
	"github.com/pburakov/playback/runner"
	"github.com/pburakov/playback/util"
)

// Playback program runner.
func main() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	log.SetOutput(os.Stdout)

	rand.Seed(time.Now().Unix())

	c := config.Init()

	in := initReader(c)
	t := initTopic(c)
	out := initOutput(t, c)

	initPlayback(in, out, c)
}

func initReader(c *config.ProgramConfig) input.FileReader {
	var r input.FileReader
	var e error
	switch c.FileType {
	case config.CSV:
		r, e = csv.Init(c.FilePath, c.TSColumn, c.TSFormat)
		break
	case config.Avro:
		r, e = avro.Init(c.FilePath, c.TSColumn, c.TSFormat)
		break
	case config.JSON:
		r, e = json.Init(c.FilePath, c.TSColumn, c.TSFormat)
		break
	default:
		e = fmt.Errorf("error initializing reader for type %q", c.FileType)
	}
	if e != nil {
		util.Fatal(e)
		return nil
	}
	return r
}

// initTopic constructs PubSub clients, verifies if given PubSub topic exists
// and constructs Topic instance.
func initTopic(c *config.ProgramConfig) *pubsub.Topic {
	ctx1, c1 := context.WithTimeout(context.Background(), c.Timeout)
	defer c1()
	p, e := pubsub.NewClient(ctx1, c.ProjectID)
	if e != nil {
		util.Fatal(e)
		return nil
	}
	t := p.Topic(c.Topic)
	ctx2, c2 := context.WithTimeout(context.Background(), c.Timeout)
	defer c2()
	if b, e := t.Exists(ctx2); e != nil || !b {
		util.Fatal(errors.New("topic does not exist or unexpected pubsub error"))
		return nil
	}
	return t
}

// initPlayback initiates configured playback mode. Log messages are printed before
// and after the playback is performed.
func initPlayback(in input.FileReader, out func(string, []byte), c *config.ProgramConfig) {
	switch c.Mode {
	case config.Instant:
		log.Printf("Starting playback in instant mode...")
		runner.PlayInstant(in, out)
	case config.Paced:
		log.Printf("Starting playback in paced mode...")
		runner.PlayPaced(in, out, c.Delay, c.MaxJitterMSec)
	case config.Relative:
		log.Printf("Starting playback in relative mode...")
		runner.PlayRelative(in, out, c.Window, c.MaxJitterMSec)
	default:
		util.Fatal(fmt.Errorf("unknown mode %d", c.Mode))
		return
	}

	log.Print("Playback stopped")
}

// initOutput returns preconfigured publishing action function.
func initOutput(t *pubsub.Topic, c *config.ProgramConfig) func(string, []byte) {
	return func(tag string, d []byte) {
		output.Publish(t, tag, d, c.Timeout)
	}
}
