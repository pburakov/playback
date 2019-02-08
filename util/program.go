package util

import (
	"log"
	"os"
)

func Fatal(e error) {
	log.Fatalf("Error: %s\n", e)
	os.Exit(1)
}
