package output

import (
	"cloud.google.com/go/pubsub"
	"context"
	"log"
	"time"
)

// Publish handles PubSub publishing procedure synchronously. Errors are ignored and logged.
func Publish(t *pubsub.Topic, tag string, d []byte, to time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), to)
	defer cancel()

	res := t.Publish(ctx, &pubsub.Message{Data: d})
	id, e := res.Get(ctx)
	if e != nil {
		log.Printf("Error publishing message: %s", e)
		return
	}
	log.Printf("Published message id %s (%s)", id, tag)
}
