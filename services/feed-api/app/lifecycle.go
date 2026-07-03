package app

import (
	"io"
	"log/slog"
)

func closeChatLifecycleCloser(closer io.Closer) {
	if closer == nil {
		return
	}
	if err := closer.Close(); err != nil {
		slog.Warn("close lifecycle resource", "err", err)
	}
}
