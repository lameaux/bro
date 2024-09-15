package app

import (
	"flag"
	"github.com/rs/zerolog/log"
)

type Flags struct {
	Args []string

	Debug        bool
	Silent       bool
	LogJson      bool
	SkipBanner   bool
	SkipResults  bool
	SkipExitCode bool
	BrodAddr     string
	Group        string
}

func ParseFlags() *Flags {
	debug := flag.Bool("debug", false, "enable debug mode")
	silent := flag.Bool("silent", false, "enable silent mode")
	logJson := flag.Bool("logJson", false, "log as json")
	skipBanner := flag.Bool("skipBanner", false, "skip banner")
	skipResults := flag.Bool("skipResults", false, "skip results")
	skipExitCode := flag.Bool("skipExitCode", false, "skip exit code")
	brodAddr := flag.String("brodAddr", "", "brod address")
	group := flag.String("group", "", "group")

	flag.Parse()

	f := &Flags{
		Debug:        *debug,
		Silent:       *silent,
		LogJson:      *logJson,
		SkipBanner:   *skipBanner,
		SkipResults:  *skipResults,
		SkipExitCode: *skipExitCode,
		BrodAddr:     *brodAddr,
		Group:        *group,

		Args: flag.Args(),
	}

	log.Debug().Any("flags", f).Msg("flags parsed")

	return f
}
