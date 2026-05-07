package turn

import "log"

func warnStore(op string, err error) {
	if err != nil {
		log.Printf("%s: %v", op, err)
	}
}
