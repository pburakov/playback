package output

import (
	"cloud.google.com/go/pubsub"
	"context"
	"github.com/stretchr/testify/assert"
	"pburakov.io/playback/test"
	"testing"
	"time"
)

const (
	testTimeout = 5 * time.Second
)

func TestPublish(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode.")
	}
	topic, sub := setup(t)

	success := make(chan bool, 1)
	go subscribe(t, sub, "foobar", success)

	Publish(topic, "baz", []byte("foobar"), testTimeout)
	waitForSuccess(t, success)
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

// waitForSuccess waits up to 5 seconds for delivery
func waitForSuccess(t *testing.T, success chan bool) {
	for i := 0; i < 5; i++ {
		select {
		case <-success:
			return
		case <-time.After(1 * time.Second):
			continue
		}
	}
	t.Log("time out waiting for message delivery")
	t.Fail()
}
