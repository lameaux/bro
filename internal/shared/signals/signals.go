package signals

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog/log"
)

func Handle(blocking bool, shutdownFn func()) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	signalReceiver := func() {
		sig := <-sigCh
		log.Info().Str("signal", sig.String()).Msg("received signal")
		shutdownFn()
	}

	if blocking {
		signalReceiver()
	} else {
		go signalReceiver()
	}
}
