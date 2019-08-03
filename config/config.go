package config

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"playback/util"
)

type Mode uint
type FileType string

const (
	Paced    Mode = 0
	Instant  Mode = 1
	Relative Mode = 2
)

const (
	CSV  FileType = "csv"
	Avro FileType = "avro"
	JSON FileType = "json"
)

const (
	DefaultTSFormat    = "2006-01-02T15:04:05.999999Z07:00"
	DefaultTimeoutMSec = 5000
	DefaultWindowMSec  = 250
	DefaultJitterMSec  = 100
	DefaultDelayMSec   = 1000
)

// ProgramConfig hold program runtime settings
type ProgramConfig struct {
	Mode          Mode
	FilePath      string
	FileType      FileType
	TSColumn      string
	TSFormat      string
	ProjectID     string
	Topic         string
	Window        time.Duration
	Timeout       time.Duration
	MaxJitterMSec int
	Delay         time.Duration
}

// InitConfig validates inputs and returns completed program configuration. Terminates on validation errors.
func Init() *ProgramConfig {
	fMode := flag.Uint("mode", 0, "Playback mode: 0 - paced, 1 - instant, 2 - relative.")
	fPath := flag.String("input", "", "Path to input file. Supported formats: JSON (newline delimited), CSV and Avro.")
	fColName := flag.String("ts_column", "", "Name of the timestamp column for relative playback mode. The input data must be sorted by that column.")
	fTSFormat := flag.String("format", DefaultTSFormat, "Timestamp format for relative playback mode. Layouts must use the reference time Mon Jan 2 15:04:05 MST 2006 to show the pattern with which to format/parse a given time/string.")
	fProjectID := flag.String("project_id", "", "Output Google Cloud project id.")
	fTopic := flag.String("topic", "", "Output PubSub topic.")
	fWindowMSec := flag.Uint("window", DefaultWindowMSec, "Event accumulation window for relative playback mode, in milliseconds. Use higher values if input event distribution on the timeline is sparse, lower values for a more dense event distribution.")
	fJitterMSec := flag.Int("jitter", DefaultJitterMSec, "Max jitter for relative and paced playback, in milliseconds.")
	fTimeoutMSec := flag.Uint("timeout", DefaultTimeoutMSec, "Publish request timeout, in milliseconds.")
	fDelayMSec := flag.Uint("delay", DefaultDelayMSec, "Delay between line reads for paced playback, in milliseconds.")
	flag.Parse()

	if len(*fProjectID) == 0 || len(*fTopic) == 0 {
		util.Fatal(errors.New("invalid project id or topic name"))
		return nil
	}

	fileType, e := validateFile(*fPath)
	if e != nil {
		util.Fatal(e)
		return nil
	}
	return &ProgramConfig{
		Mode:          Mode(*fMode),
		FilePath:      *fPath,
		FileType:      fileType,
		TSColumn:      *fColName,
		TSFormat:      *fTSFormat,
		ProjectID:     *fProjectID,
		Topic:         *fTopic,
		Window:        time.Duration(*fWindowMSec * 1000000),
		Timeout:       time.Duration(*fTimeoutMSec * 1000000),
		Delay:         time.Duration(*fDelayMSec * 1000000),
		MaxJitterMSec: *fJitterMSec,
	}
}

// validateFile checks if file exists and validates file extension
func validateFile(f string) (FileType, error) {
	if len(f) == 0 {
		return "", errors.New("no input file")
	}
	if _, err := os.Stat(f); os.IsNotExist(err) {
		return "", fmt.Errorf("file %q does not exist", f)
	}
	ext := ""
	for _, s := range strings.Split(f, ".") {
		ext = s
	}
	switch ext {
	case string(CSV), string(Avro), string(JSON):
		ft := FileType(ext)
		return ft, nil
	default:
		return "", errors.New("unsupported file type")
	}
}
