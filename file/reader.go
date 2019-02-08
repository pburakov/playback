package file

import "time"

type Reader interface {
	ReadLine() (ts time.Time, data []byte, err error)
	// TODO: add PeekNextTS() method
}
