package logutil

import "log"

func WarnIfErr(op string, err error) {
	if err != nil {
		log.Printf("%s: %v", op, err)
	}
}
