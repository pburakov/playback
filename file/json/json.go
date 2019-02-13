package json

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"playback/file"
	"playback/util"
	"time"
)

type properties struct {
	tsColumn string
	tsFormat string
}

type JSONReader struct {
	r *bufio.Reader
	p *properties
}

var _ file.Reader = (*JSONReader)(nil)

func Init(path string, colName string, tsFormat string) (*JSONReader, error) {
	f, e := os.Open(path)
	if e != nil {
		return nil, e
	}
	r := bufio.NewReader(f)
	return &JSONReader{r: r, p: &properties{tsColumn: colName, tsFormat: tsFormat}}, nil
}

func (j *JSONReader) ReadLine() (time.Time, []byte, error) {
	line, _, e := j.r.ReadLine()
	if e != nil {
		return util.DefaultTimestamp(), nil, e
	}
	m := make(map[string]interface{})
	if e := json.Unmarshal(line, &m); e != nil {
		return util.DefaultTimestamp(), nil, e
	}
	if v, found := m[j.p.tsColumn]; !found {
		return util.DefaultTimestamp(), nil, fmt.Errorf("invalid timestamp column %q", j.p.tsColumn)
	} else {
		if t, ok := v.(string); !ok {
			return util.DefaultTimestamp(), nil, fmt.Errorf("unexpected timestamp field %q", v)
		} else {
			ts, e := time.Parse(j.p.tsFormat, t)
			return ts, line, e
		}
	}
}
