package json

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

const (
	testFile     = "input_test.json"
	testColumn   = "bar"
	testTSFormat = "2006-01-02T15:04:05.999999"
)

func TestInitAndReadLine(t *testing.T) {
	r, e := Init(testFile, testColumn, testTSFormat)

	assert.NoError(t, e)
	assert.Equal(t, &properties{
		tsColumn: testColumn,
		tsFormat: testTSFormat,
	}, r.p)

	ts, l, e := r.ReadLine()

	expected := `{"foo":"1","bar":"2019-02-11T15:20:09.514626","baz":["1","2","3"],"faz":{"A":"foo","B":42.42}}`

	assert.NoError(t, e)
	assert.Equal(t, time.Date(2019, 02, 11, 15, 20, 9, 514626000, time.UTC), ts)
	assert.Equal(t, expected, string(l))
}

func TestInitErrors(t *testing.T) {
	r, e := Init("non_existent_file", testColumn, testTSFormat)

	assert.Error(t, e)
	assert.Nil(t, r)

	// JSON schema compliance isn't validated until the first read
	r, _ = Init(testFile, "non_existent_column", testTSFormat)
	_, _, e = r.ReadLine()

	assert.Error(t, e)

	r, _ = Init(testFile, testColumn, "bad_format")
	_, _, e = r.ReadLine()

	assert.Error(t, e)
}
