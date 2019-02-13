package csv

import (
	"github.com/stretchr/testify/assert"
	"io"
	"testing"
	"time"
)

const (
	testFile     = "input_test.csv"
	testColumn   = "baz"
	testTSFormat = "2006-01-02 15:04:05 UTC"
)

func TestInitAndReadLines(t *testing.T) {
	r, e := Init(testFile, testColumn, testTSFormat)

	assert.NoError(t, e)
	assert.Equal(t, &properties{
		headers:  []string{"foo", "bar", "baz"},
		tsFormat: testTSFormat,
		tsColumn: testColumn,
	}, r.p)

	ts, l, e := r.ReadLineWithTS()

	assert.NoError(t, e)
	assert.Equal(t, time.Date(2019, 02, 04, 21, 16, 19, 0, time.UTC), ts)
	assert.Equal(t, `{"bar":"2","baz":"2019-02-04 21:16:19 UTC","foo":"1"}`, string(l))

	l, e = r.ReadLine()

	assert.NoError(t, e)
	assert.Equal(t, `{"bar":"4","baz":"2019-02-07 12:53:31 UTC","foo":"3"}`, string(l))

	_, e = r.ReadLine()

	assert.Equal(t, io.EOF, e)
}

func TestInitErrors(t *testing.T) {
	r, e := Init("non_existent_file", testColumn, testTSFormat)

	assert.Error(t, e)
	assert.Nil(t, r)
}

func TestTSErrors(t *testing.T) {
	r, _ := Init(testFile, "non_existing_column", testTSFormat)
	_, _, e := r.ReadLineWithTS()

	assert.Error(t, e)

	r, _ = Init(testFile, testColumn, "bad format")
	_, _, e = r.ReadLineWithTS()

	assert.Error(t, e)
}
