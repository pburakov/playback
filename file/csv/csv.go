package csv

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"playback/file"
	"playback/util"
	"time"
)

type properties struct {
	headers  []string
	tsColumn int
	tsFormat string
}

type CSVReader struct {
	r *csv.Reader
	p *properties
}

var _ file.Reader = (*CSVReader)(nil)

func Init(path string, colName string, tsFormat string) (*CSVReader, error) {
	f, e := os.Open(path)
	if e != nil {
		return nil, e
	}

	log.Printf("Loading csv file %q", path)

	r := csv.NewReader(bufio.NewReader(f))

	line, e := r.Read()
	if e != nil {
		return nil, e
	}

	t, e := parseSchema(line, colName)
	if e != nil {
		return nil, e
	}

	log.Printf("Using column %q (%d) as timestamp column", colName, t)

	return &CSVReader{r: r, p: &properties{headers: line, tsColumn: t, tsFormat: tsFormat}}, nil
}

// ReadLine returns CSV entry as a serialized JSON k-v object.
func (c *CSVReader) ReadLine() (ts time.Time, data []byte, e error) {
	row, e := c.r.Read()
	if e != nil {
		return util.DefaultTimestamp(), nil, e
	}

	// Extract timestamp
	v := row[c.p.tsColumn]
	ts, e = time.Parse(c.p.tsFormat, v)
	if e != nil {
		return util.DefaultTimestamp(), nil, e
	}

	// Build map
	rec := make(map[string]string)
	for i, h := range c.p.headers {
		rec[h] = row[i]
	}

	// Convert to json
	data, e = json.Marshal(rec)
	if e != nil {
		return util.DefaultTimestamp(), nil, fmt.Errorf("unable to convert csv record to json: %s", e)
	}
	return ts, data, e
}

// parseSchema parses header and returns serial number of a timestamp column
func parseSchema(headers []string, col string) (int, error) {
	t := -1
	for i, c := range headers {
		if c == col {
			t = i
		}
	}
	if t == -1 {
		return -1, fmt.Errorf("invalid timestamp column %q", col)
	}
	return t, nil
}
