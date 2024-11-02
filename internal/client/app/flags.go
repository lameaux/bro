package app

import (
	"flag"
	"time"

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

	URL      bool
	Method   string
	RPS      int
	Threads  int
	Duration time.Duration
	Timeout  time.Duration
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

	// to run scenarios without config
	url := flag.Bool("url", false, "url")
	flag.BoolVar(url, "u", *url, "alias for url")

	rps := flag.Int("rps", 0, "rps")
	flag.IntVar(rps, "r", *rps, "alias for rps")

	threads := flag.Int("threads", 0, "threads")
	flag.IntVar(threads, "t", *threads, "alias for threads")

	duration := flag.Duration("duration", 0, "duration")
	flag.DurationVar(duration, "d", *duration, "alias for duration")

	method := flag.String("method", "", "method")
	flag.StringVar(method, "m", *method, "alias for method")

	timeout := flag.Duration("timeout", 0, "timeout")

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

		URL:      *url,
		RPS:      *rps,
		Threads:  *threads,
		Duration: *duration,
		Timeout:  *timeout,
		Method:   *method,

		Args: flag.Args(),
	}

	if flags.Debug {
		log.Debug().Any("flags", flags).Msg("flags parsed")
	}

	return flags
}
