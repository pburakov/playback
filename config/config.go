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
	Instant  Mode = 0
	Paced    Mode = 1
	Relative Mode = 2
)

const (
	CSV  FileType = "csv"
	Avro FileType = "avro"
	JSON FileType = "json"
)

const (
	DefaultTSFormat    = "2006-01-02T15:04:05.999999"
	DefaultTimeoutMSec = 5000
	DefaultWindowMSec  = 250
	DefaultJitterMSec  = 150
	DefaultDelayMSec   = 500
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
	fMode := flag.Int("m", 0, "Mode: 0 - instant (default), 1 - paced, 2 - relative")
	fPath := flag.String("i", "", "Path to input file")
	fColName := flag.String("c", "", "Name of the timestamp column. The input data must be sorted by that column")
	fTSFormat := flag.String("f", DefaultTSFormat, "Timestamp format")
	fProjectID := flag.String("p", "", "Google Cloud project id")
	fTopic := flag.String("t", "", "Output topic")
	fWindowMSec := flag.Int("w", DefaultWindowMSec, "Event accumulation window for relative playback (milliseconds)")
	fJitterMSec := flag.Int("j", DefaultJitterMSec, "Max jitter for relative and paced playback (milliseconds)")
	fTimeoutMSec := flag.Int("o", DefaultTimeoutMSec, "Publish request timeout (milliseconds)")
	fDelayMSec := flag.Int("d", DefaultDelayMSec, "Delay between publish requests for paced playback (milliseconds)")
	flag.Parse()

	if len(*fProjectID) == 0 || len(*fTopic) == 0 {
		util.Fatal(errors.New("invalid project id or topic name"))
	}

	if len(*fColName) == 0 {
		util.Fatal(errors.New("invalid or empty column value"))
	}

	fileType, e := validateFile(*fPath)
	if e != nil {
		util.Fatal(e)
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
