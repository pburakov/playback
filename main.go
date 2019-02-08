package main

import (
	"cloud.google.com/go/pubsub"
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"playback/config"
	"playback/file"
	"playback/file/avro"
	"playback/file/csv"
	"playback/file/json"
	"playback/runner"
	"playback/util"
	"time"
)

// PubSub playback program
//
// Example usage:
// ./playback -i test_data.csv -p spotify-payment-risk-sandbox -c Creation_Date -f '2006-01-02 15:04:05 UTC' -t billing-events-sandbox
func main() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	rand.Seed(time.Now().Unix())

	c := config.Init()

	r := initReader(c)
	t := initTopic(c)

	initPlayback(r, runner.Output(t, c), c)
}

func initReader(c *config.ProgramConfig) file.Reader {
	var r file.Reader
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
	}
	return r
}

func initTopic(c *config.ProgramConfig) *pubsub.Topic {
	ctx1, c1 := context.WithTimeout(context.Background(), c.Timeout)
	defer c1()
	p, e := pubsub.NewClient(ctx1, c.ProjectID)
	if e != nil {
		util.Fatal(e)
	}
	t := p.Topic(c.Topic)
	ctx2, c2 := context.WithTimeout(context.Background(), c.Timeout)
	defer c2()
	if b, e := t.Exists(ctx2); e != nil || !b {
		util.Fatal(errors.New("topic does not exist or unexpected pubsub error"))
	}
	return t
}

func initPlayback(in file.Reader, action func(time.Time, []byte), c *config.ProgramConfig) {
	log.Print("Starting playback...")

	switch c.Mode {
	case config.Instant:
		log.Printf("Initilized instant playback mode")
		runner.PlayInstant(in, action)
	case config.Paced:
		log.Printf("Initilized paced playback mode")
		runner.PlayPaced(in, action, c.Delay, c.MaxJitterMSec)
	case config.Relative:
		log.Printf("Initilized relative playback mode")
		runner.PlayRelative(in, action, c.Window, c.MaxJitterMSec)
	default:
		util.Fatal(fmt.Errorf("unknown mode %d", c.Mode))
	}

	log.Print("Playback stopped")
}
