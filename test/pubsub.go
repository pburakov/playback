package test

import (
	"cloud.google.com/go/pubsub"
	"context"
	"fmt"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"log"
	"net/http"
	"os"
	"time"
)

// PubsubTestContainer holds pubsub client for instantiated PubSub test-container
type PubsubTestContainer struct {
	PubSubClient *pubsub.Client
}

// Implementing waiting strategy interface
var _ wait.Strategy = (*pubsubWaitingStrategy)(nil)

const (
	pubsubPort = "6379"
	projectId  = "test-project"
)

type pubsubWaitingStrategy struct {
	// max time to probe for a successful connection once container have started
	probingTimeout time.Duration

	// all Strategies should have a startupTimeout to avoid waiting infinitely
	startupTimeout time.Duration
}

func (ws *pubsubWaitingStrategy) WaitUntilReady(ctx context.Context, target wait.StrategyTarget) (err error) {
	// limit context to startupTimeout
	ctx, cancelContext := context.WithTimeout(ctx, ws.startupTimeout)
	defer cancelContext()

	ipAddress, err := target.Host(ctx)
	if err != nil {
		return
	}
	port, err := target.MappedPort(ctx, pubsubPort)
	if err != nil {
		return
	}
	url := fmt.Sprintf("http://%s:%d/v1/projects/%s/topics", ipAddress, port.Int(), projectId)

	client := http.Client{Timeout: ws.startupTimeout}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req = req.WithContext(ctx)

	for i := 0.0; i < ws.probingTimeout.Seconds(); i++ {
		resp, err := client.Do(req)
		if err != nil {
			time.Sleep(1 * time.Second)
			log.Print("waiting for pubsub container to start")
			continue
		}
		if resp.StatusCode == http.StatusOK {
			log.Print("connected to pubsub test-container")
			return nil
		}
	}
	return fmt.Errorf("failed probing pubsub container")
}

// BindPubSub constructs new PubSub test-container instance with a bound pubsub.Client
func BindPubSub() *PubsubTestContainer {
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "google/cloud-sdk:latest",
		Cmd:          fmt.Sprintf("gcloud --quiet --project %s beta emulators pubsub start --host-port 0.0.0.0:%s", projectId, pubsubPort),
		ExposedPorts: []string{fmt.Sprintf("%s/tcp", pubsubPort)},
		WaitingFor: &pubsubWaitingStrategy{
			probingTimeout: 10 * time.Second,
			startupTimeout: 60 * time.Second,
		},
	}

	pubsubC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		panic(err)
	}

	host, err := pubsubC.Host(ctx)
	if err != nil {
		panic(err)
	}
	mp, err := pubsubC.MappedPort(ctx, pubsubPort)
	if err != nil {
		panic(err)
	}

	err = os.Setenv("PUBSUB_EMULATOR_HOST", fmt.Sprintf("%s:%d", host, mp.Int()))
	if err != nil {
		panic(err)
	}
	defer os.Unsetenv("PUBSUB_EMULATOR_HOST")

	psc, err := pubsub.NewClient(ctx, projectId)
	if err != nil {
		panic(err)
	}
	return &PubsubTestContainer{
		PubSubClient: psc,
	}
}
