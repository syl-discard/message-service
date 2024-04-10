package logger

import (
	"log"
	"os"
)

var (
	WARN  = log.New(os.Stderr, "[WARNING]\t", log.LstdFlags|log.Lmsgprefix)
	ERROR = log.New(os.Stderr, "[ERROR]\t", log.LstdFlags|log.Lmsgprefix)
	LOG   = log.New(os.Stdout, "[INFO]\t", log.LstdFlags|log.Lmsgprefix)
)

func FailOnError(err error, msg string) {
	if err != nil {
		ERROR.Printf("%s: %s\n", msg, err)
	}
}
