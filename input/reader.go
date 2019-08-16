package input

import "time"

type FileReader interface {
	// ReadLineWithTS reads the next line from the input file, extracts the timestamp
	// value and returns it with the binary data from the input line.
	ReadLineWithTS() (ts time.Time, data []byte, e error)

	// ReadLine reads the next line from the input file and returns the binary data.
	ReadLine() (data []byte, e error)

	// TODO: add PeekNextTS() method
}
