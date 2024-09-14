package app

import (
	"fmt"
	"github.com/lameaux/bro/internal/shared/banner"
	"github.com/rs/zerolog/log"
)

func (a *App) printAbout() {
	if !a.flags.Silent && !a.flags.SkipBanner {
		fmt.Print(banner.Banner)
	}

	log.Info().
		Str("version", a.appVersion).
		Str("build", a.appBuild).
		Msg(a.appName)
}
