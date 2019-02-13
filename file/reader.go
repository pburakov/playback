package file

import "time"

type Reader interface {
	// ReadLine reads the next line from the input file, extracts the timestamp
	// value and returns it with the binary data from the input line.
	ReadLine() (ts time.Time, data []byte, e error)
	// TODO: add PeekNextTS() method
}
