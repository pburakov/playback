// A suite of basic end-to-end tests to verify workflows
package runner

import (
	"cloud.google.com/go/pubsub"
	"context"
	"github.com/stretchr/testify/assert"
	"playback/config"
	"playback/file/json"
	"playback/test"
	"testing"
	"time"
)

var testConfig = &config.ProgramConfig{
	Timeout: 5 * time.Second,
}

const (
	testFile     = "runner_test.json"
	testColumn   = "ts"
	testTSFormat = "2006-01-02T15:04:05.999999"
	testWindow   = 100 * time.Millisecond
	testDelay    = 500 * time.Millisecond
	testJitter   = 50
	testTimeout  = 5 * time.Second
)

func TestPublish(t *testing.T) {
	topic, sub := setup(t)

	success := make(chan bool, 1)
	go subscribe(t, sub, "foobar", success)

	publish(topic, "baz", []byte("foobar"), testTimeout)
	waitForDelivery(t, success)
}

func TestPlayInstant(t *testing.T) {
	topic, sub := setup(t)

	success := make(chan bool, 1)
	go subscribe(t, sub, `{"ts":"2019-02-11T15:20:09.514626","val":"foo"}`, success)

	in, e := json.Init(testFile, testColumn, testTSFormat)
	assert.NoError(t, e)

	PlayInstant(in, Output(topic, testConfig))

	waitForDelivery(t, success)
}

func TestPlayRelative(t *testing.T) {
	topic, sub := setup(t)

	success := make(chan bool, 1)
	go subscribe(t, sub, `{"ts":"2019-02-11T15:20:09.514626","val":"foo"}`, success)

	in, e := json.Init(testFile, testColumn, testTSFormat)
	assert.NoError(t, e)

	PlayRelative(in, Output(topic, testConfig), testWindow, testJitter)

	waitForDelivery(t, success)
}

func TestPlayPaced(t *testing.T) {
	topic, sub := setup(t)

	success := make(chan bool, 1)
	go subscribe(t, sub, `{"ts":"2019-02-11T15:20:09.514626","val":"foo"}`, success)

	in, e := json.Init(testFile, testColumn, testTSFormat)
	assert.NoError(t, e)

	PlayPaced(in, Output(topic, testConfig), testDelay, testJitter)

	waitForDelivery(t, success)
}

func setup(t *testing.T) (*pubsub.Topic, *pubsub.Subscription) {
	ps := test.BindPubSub().PubSubClient

	topic, e := ps.CreateTopic(context.Background(), "test-topic")
	assert.NoError(t, e)

	sub, e := ps.CreateSubscription(context.Background(), "test-sub", pubsub.SubscriptionConfig{Topic: topic})
	assert.NoError(t, e)

	return topic, sub
}

func subscribe(t *testing.T, sub *pubsub.Subscription, expected string, success chan bool) {
	ctx := context.Background()
	e := sub.Receive(ctx, func(ctx context.Context, m *pubsub.Message) {
		m.Ack()
		assert.Equal(t, expected, string(m.Data))
		success <- true
	})
	if e != nil {
		t.Error(e)
	}
}

// waitForDelivery waits up to 5 seconds for delivery
func waitForDelivery(t *testing.T, success chan bool) {
	for i := 0; i < 5; i++ {
		select {
		case <-success:
			return
		case <-time.After(1 * time.Second):
			continue
		}
	}
	t.Fail()
}
