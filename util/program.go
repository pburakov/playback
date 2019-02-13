package util

import (
	"log"
	"os"
)

// Fatal prints out error message and quits with exit code 1.
func Fatal(e error) {
	log.Fatalf("Error: %s\n", e)
	os.Exit(1)
}
