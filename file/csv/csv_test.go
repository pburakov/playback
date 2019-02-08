package csv

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

const (
	testFile         = "input_test.csv"
	testColumn       = "baz"
	testColumnSerial = 2
	testTSFormat     = "2006-01-02 15:04:05 UTC"
)

func TestInitAndReadLine(t *testing.T) {
	r, e := Init(testFile, testColumn, testTSFormat)

	assert.NoError(t, e)
	assert.Equal(t, &properties{
		headers:  []string{"foo", "bar", "baz"},
		tsColumn: testColumnSerial,
		tsFormat: testTSFormat,
	}, r.p)

	ts, l, e := r.ReadLine()

	assert.NoError(t, e)
	assert.Equal(t, time.Date(2019, 02, 04, 21, 16, 19, 0, time.UTC), ts)
	assert.Equal(t, `{"bar":"2","baz":"2019-02-04 21:16:19 UTC","foo":"1"}`, string(l))
}

func TestInitErrors(t *testing.T) {
	r, e := Init(testFile, "non_existing_column", testTSFormat)

	assert.Error(t, e)
	assert.Nil(t, r)

	r, e = Init("non_existent_file", testColumn, testTSFormat)

	assert.Error(t, e)
	assert.Nil(t, r)

	r, _ = Init(testFile, testColumn, "bad format")
	_, _, e = r.ReadLine()

	assert.Error(t, e)
}
