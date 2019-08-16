// A suite of basic end-to-end tests to verify workflows
package runner

import (
	"io"
	"log"
	"playback/input"
	"testing"
	"time"
)

const (
	expectedTsFormat = "2006-01-02T15:04:05.999999Z07:00"
	expectedTs       = "2019-02-11T15:20:09.514626Z"
	expectedPayload  = `{"ts":"2019-02-11T15:20:09.514626Z","val":"foo"}`
	testWindow       = 100 * time.Millisecond
	testDelay        = 500 * time.Millisecond
	testJitter       = 50
)

func TestPlayInstant(t *testing.T) {
	success := make(chan bool, 1)

	in := initTestReader(t)
	PlayInstant(in, testOutput(expectedPayload, success))
	waitForSuccess(t, success)
}

func TestPlayRelative(t *testing.T) {
	success := make(chan bool, 1)

	in := initTestReader(t)
	PlayRelative(in, testOutput(expectedPayload, success), testWindow, testJitter)
	waitForSuccess(t, success)
}

func TestPlayPaced(t *testing.T) {
	success := make(chan bool, 1)

	in := initTestReader(t)
	PlayPaced(in, testOutput(expectedPayload, success), testDelay, testJitter)
	waitForSuccess(t, success)
}

func testOutput(expected string, success chan bool) func(string, []byte) {
	return func(s string, b []byte) {
		if string(b) == expected {
			log.Printf("published test message (%s)", s)
			success <- true
		}
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
	t.Log("time out waiting for success")
	t.Fail()
}

type testReader struct {
	t       *testing.T
	payload []byte
}

var _ input.FileReader = (*testReader)(nil)

func initTestReader(t *testing.T) *testReader {
	r := new(testReader)
	r.payload = []byte(expectedPayload)
	r.t = t
	return r
}

func (r *testReader) ReadLineWithTS() (ts time.Time, data []byte, e error) {
	if r.payload != nil {
		t, _ := time.Parse(expectedTsFormat, expectedTs)
		p := r.payload
		r.payload = nil // do not send payload on next invocation
		return t, p, nil
	} else {
		return time.Now(), nil, io.EOF
	}
}

func (r *testReader) ReadLine() (data []byte, e error) {
	if r.payload != nil {
		p := r.payload
		r.payload = nil // do not send payload on next invocation
		return p, nil
	} else {
		return nil, io.EOF
	}
}
