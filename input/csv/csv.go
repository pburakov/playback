package csv

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/pburakov/playback/input"
	"github.com/pburakov/playback/util"
)

type properties struct {
	headers  []string
	tsColumn string
	tsFormat string
}

type CSVReader struct {
	r *csv.Reader
	p *properties
}

var _ input.FileReader = (*CSVReader)(nil)

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

	return &CSVReader{r: r, p: &properties{headers: line, tsColumn: colName, tsFormat: tsFormat}}, nil
}

// ReadLine returns CSV entry as a serialized JSON k-v object.
func (c *CSVReader) ReadLineWithTS() (ts time.Time, data []byte, e error) {
	row, e := c.r.Read()
	if e != nil {
		return util.DefaultTimestamp(), nil, e
	}

	// Build map
	rec := make(map[string]string)
	for i, h := range c.p.headers {
		rec[h] = row[i]
	}

	ts, e = extractTimestamp(rec, c.p.tsColumn, c.p.tsFormat)
	if e != nil {
		return util.DefaultTimestamp(), nil, e
	}

	// Convert to json
	data, e = json.Marshal(rec)
	if e != nil {
		return util.DefaultTimestamp(), nil, fmt.Errorf("unable to convert csv record to json: %s", e)
	}
	return ts, data, e
}

func (c *CSVReader) ReadLine() (data []byte, e error) {
	row, e := c.r.Read()
	if e != nil {
		return nil, e
	}

	// Build map
	rec := make(map[string]string)
	for i, h := range c.p.headers {
		rec[h] = row[i]
	}

	// Convert to json
	data, e = json.Marshal(rec)
	if e != nil {
		return nil, fmt.Errorf("unable to convert csv record to json: %s", e)
	}
	return data, e
}

// extractTimestamp extracts timestamp from a mapped row.
func extractTimestamp(row map[string]string, col string, format string) (time.Time, error) {
	if v, found := row[col]; !found {
		return util.DefaultTimestamp(), fmt.Errorf("invalid timestamp column %q", col)
	} else {
		ts, e := time.Parse(format, v)
		if e != nil {
			return util.DefaultTimestamp(), e
		}
		return ts, nil
	}
}
