package app

import (
	"fmt"
	"runtime"
	"runtime/debug"

	"github.com/lameaux/bro/internal/shared/banner"
	"github.com/rs/zerolog/log"
)

func (a *App) printAbout() {
	if !a.flags.Silent && !a.flags.SkipBanner {
		fmt.Print(banner.Banner) //nolint:forbidigo
	}

	log.Info().
		Str("version", a.appVersion).
		Str("build", a.appBuild).
		Int("cpus", runtime.NumCPU()).
		Int("GOMAXPROCS", runtime.GOMAXPROCS(-1)).
		Msg(a.appName)

	if a.flags.BuildInfo {
		logBuildInfo()
	}
}

func logBuildInfo() {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		log.Warn().Msg("failed to read build info")

		return
	}

	log.Info().
		Any("info", info).
		Msg("build")
}
