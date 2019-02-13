package avro

import (
	"github.com/stretchr/testify/assert"
	"io"
	"testing"
	"time"
)

const (
	testFile           = "input_test.avro"
	testColumn         = "bar"
	testDateTimeFormat = "2006-01-02T15:04:05.999999"
)

func TestInitAndReadLines(t *testing.T) {
	r, e := Init(testFile, testColumn, testDateTimeFormat)

	assert.NoError(t, e)
	assert.Equal(t, &properties{
		tsColumn: testColumn,
		tsFormat: testDateTimeFormat,
	}, r.p)

	ts, _, e := r.ReadLineWithTS()

	assert.NoError(t, e)
	assert.Equal(t, time.Date(2019, 02, 11, 15, 20, 9, 514626000, time.UTC), ts)

	l, e := r.ReadLine()

	assert.NoError(t, e)
	assert.NotNil(t, l)

	_, e = r.ReadLine()

	assert.Equal(t, io.EOF, e)
}

func TestTSInput(t *testing.T) {
	// File ts_test.avro contains columns of DATETIME (bar) and TIMESTAMP (baz) types
	r, _ := Init("ts_test.avro", "bar", "2006-01-02T15:04:05.999999")

	ts, _, e := r.ReadLineWithTS()

	assert.NoError(t, e)
	assert.Equal(t, time.Date(2019, 02, 11, 17, 56, 42, 53944000, time.UTC), ts)

	// TIMESTAMP field is internally represented as long microseconds type, so format is ignored.
	r, _ = Init("ts_test.avro", "baz", "doesn't matter")

	ts, _, e = r.ReadLineWithTS()

	assert.NoError(t, e)
	assert.Equal(t, time.Date(2019, 02, 11, 17, 56, 42, 53944000, time.UTC), ts)
}

func TestInitErrors(t *testing.T) {
	r, e := Init("non_existent_file", testColumn, testDateTimeFormat)

	assert.Error(t, e)
	assert.Nil(t, r)
}

func TestTSErrors(t *testing.T) {
	// Avro schema compliance isn't validated until the first read
	r, _ := Init(testFile, "non_existent_column", testDateTimeFormat)
	_, _, e := r.ReadLineWithTS()

	assert.Error(t, e)

	r, _ = Init(testFile, testColumn, "bad_format")
	_, _, e = r.ReadLineWithTS()

	assert.Error(t, e)
}
