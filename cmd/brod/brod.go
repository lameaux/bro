package main

import "github.com/rs/zerolog/log"

const (
	appName    = "brod"
	appVersion = "v0.0.1"
)

var GitHash string

func main() {
	log.Info().Str("version", appVersion).Str("build", GitHash).Msg(appName)

}
