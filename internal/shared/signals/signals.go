package signals

import (
	"github.com/rs/zerolog/log"
	"os"
	"os/signal"
	"syscall"
)

func Handle(blocking bool, shutdownFn func()) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	f := func() {
		sig := <-sigCh
		log.Info().Str("signal", sig.String()).Msg("received signal")
		shutdownFn()
	}

	if blocking {
		f()
	} else {
		go f()
	}
}
