package app

import (
	"flag"

	"github.com/rs/zerolog/log"
)

type Flags struct {
	Args []string

	Debug        bool
	Silent       bool
	LogJSON      bool
	SkipBanner   bool
	SkipResults  bool
	SkipExitCode bool
	BrodAddr     string
	Group        string
}

func ParseFlags() *Flags {
	debug := flag.Bool("debug", false, "enable debug mode")
	silent := flag.Bool("silent", false, "enable silent mode")
	logJSON := flag.Bool("logJson", false, "log as json")
	skipBanner := flag.Bool("skipBanner", false, "skip banner")
	skipResults := flag.Bool("skipResults", false, "skip results")
	skipExitCode := flag.Bool("skipExitCode", false, "skip exit code")
	brodAddr := flag.String("brodAddr", "", "brod address")
	group := flag.String("group", "", "group")

	flag.Parse()

	flags := &Flags{
		Debug:        *debug,
		Silent:       *silent,
		LogJSON:      *logJSON,
		SkipBanner:   *skipBanner,
		SkipResults:  *skipResults,
		SkipExitCode: *skipExitCode,
		BrodAddr:     *brodAddr,
		Group:        *group,

		Args: flag.Args(),
	}

	log.Debug().Any("flags", flags).Msg("flags parsed")

	return flags
}
