package avro

import (
	"fmt"
	"github.com/linkedin/goavro"
	"io"
	"log"
	"os"
	"pburakov.io/playback/input"
	"pburakov.io/playback/util"
	"time"
)

type properties struct {
	tsColumn string
	tsFormat string
}

type AvroReader struct {
	r *goavro.OCFReader
	p *properties
}

var _ input.FileReader = (*AvroReader)(nil)

func Init(path string, colName string, tsFormat string) (*AvroReader, error) {
	f, e := os.Open(path)
	if e != nil {
		return nil, e
	}

	r, e := goavro.NewOCFReader(f)
	if e != nil {
		return nil, e
	}

	log.Printf("Loading avro file %q (compression algorithm %q)", path, r.CompressionName())

	return &AvroReader{r: r, p: &properties{tsColumn: colName, tsFormat: tsFormat}}, nil
}

func (a *AvroReader) ReadLineWithTS() (ts time.Time, data []byte, e error) {
	s := a.r.Scan()
	if s == false {
		return util.DefaultTimestamp(), nil, io.EOF
	}
	r, e := a.r.Read()
	if e != nil {
		return util.DefaultTimestamp(), nil, e
	}
	if rec, ok := r.(map[string]interface{}); !ok {
		return util.DefaultTimestamp(), nil, fmt.Errorf("unable to parse record")
	} else {
		ts, e := extractTimestamp(rec, a.p.tsColumn, a.p.tsFormat)
		if e != nil {
			return ts, nil, e
		}
		data, e := a.r.Codec().BinaryFromNative(nil, r)
		return ts, data, e
	}
}

func (a *AvroReader) ReadLine() (data []byte, e error) {
	s := a.r.Scan()
	if s == false {
		return nil, io.EOF
	}
	r, e := a.r.Read()
	if e != nil {
		return nil, e
	}
	return a.r.Codec().BinaryFromNative(nil, r)
}

// extractTimestamp makes best guess about timestamp type and deserializes it.
func extractTimestamp(m map[string]interface{}, col string, format string) (time.Time, error) {
	ets := time.Unix(0, 0)
	if val, found := m[col]; !found {
		return ets, fmt.Errorf("timestamp column %q not found", val)
	} else {
		if mval, ok := val.(map[string]interface{}); !ok {
			return ets, fmt.Errorf("unexpected schema %q", val)
		} else {
			// Extract known timestamp serializations (TIMESTAMP and DATETIME)
			for _, v := range mval {
				switch t := v.(type) {
				case int64:
					// Microseconds to unix time
					return time.Unix(t/1000000, 1000*(t%1000000)).UTC(), nil
				case string:
					return time.Parse(format, t)
				default:
					return ets, fmt.Errorf("invalid timestamp format %q", t)
				}
			}
			return ets, fmt.Errorf("no known timestamp fields in %q", mval)
		}
	}
}
