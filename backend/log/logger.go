package log

import (
	"fmt"
	"log"
)

type logWriter struct {
}

func (writer logWriter) Write(bytes []byte) (int, error) {
	return fmt.Print(string(bytes))
}

func EnableStdOutLog() {
	log.SetFlags(0)
	log.SetOutput(new(logWriter))
}
