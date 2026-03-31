package app

import (
	"io"
	"log"
)

func closeChatLifecycleCloser(closer io.Closer) {
	if closer == nil {
		return
	}
	if err := closer.Close(); err != nil {
		log.Printf("close lifecycle resource: %v", err)
	}
}
