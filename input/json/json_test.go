package json

import (
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const (
	testFile     = "input_test.json"
	testColumn   = "bar"
	testTSFormat = "2006-01-02T15:04:05.999999"
)

func TestInitAndReadLines(t *testing.T) {
	r, e := Init(testFile, testColumn, testTSFormat)

	assert.NoError(t, e)
	assert.Equal(t, &properties{
		tsColumn: testColumn,
		tsFormat: testTSFormat,
	}, r.p)

	ts, l, e := r.ReadLineWithTS()

	expected := `{"foo":"1","bar":"2019-02-11T15:20:09.514626","baz":["1","2","3"],"faz":{"A":"foo","B":42.42}}` + "\n"

	assert.NoError(t, e)
	assert.Equal(t, time.Date(2019, 02, 11, 15, 20, 9, 514626000, time.UTC), ts)
	assert.Equal(t, expected, string(l))

	l, e = r.ReadLine()

	expected = `{"foo":"2","bar":"2019-02-06T02:22:47.327394","baz":["4","5","6"],"faz":{"A":"moo","B":43.43}}` + "\n"

	assert.NoError(t, e)
	assert.Equal(t, expected, string(l))

	_, e = r.ReadLine()

	assert.Equal(t, io.EOF, e)
}

func TestInitErrors(t *testing.T) {
	r, e := Init("non_existent_file", testColumn, testTSFormat)

	assert.Error(t, e)
	assert.Nil(t, r)
}

func TestTSErrors(t *testing.T) {
	// JSON schema compliance isn't validated until the first read
	r, _ := Init(testFile, "non_existent_column", testTSFormat)
	_, _, e := r.ReadLineWithTS()

	assert.Error(t, e)

	r, _ = Init(testFile, testColumn, "bad_format")
	_, _, e = r.ReadLineWithTS()

	assert.Error(t, e)
}
