package app

import (
	"flag"
	"time"

	"github.com/rs/zerolog/log"
)

type Flags struct {
	Args []string

	Debug      bool
	Silent     bool
	LogJSON    bool
	SkipBanner bool

	Output string
	Format string

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

func ParseFlags() *Flags { //nolint:funlen
	debug := flag.Bool("debug", false, "set log level to DEBUG")
	silent := flag.Bool("silent", false, "set log level to ERROR")
	logJSON := flag.Bool("logJson", false, "set log output format as JSON")
	skipBanner := flag.Bool("skipBanner", false, "do not show banner on start up")
	skipExitCode := flag.Bool("skipExitCode", false, "do not set exit code on test failure")
	brodAddr := flag.String("brodAddr", "", "address (host:port) of brod, e.g. brod:8080")
	group := flag.String("group", "", "test group identifier")

	// test results
	output := flag.String("output", "stdout", "output: stdout, path/to/file")
	flag.StringVar(output, "o", *output, "alias for output")

	format := flag.String("format", "txt", "format: txt or csv")
	flag.StringVar(format, "f", *format, "alias for format")

	// to run scenarios without config
	url := flag.Bool("url", false, "target URL for scenario")
	flag.BoolVar(url, "u", *url, "alias for url")

	rps := flag.Int("rps", 0, "target RPS for scenario")
	flag.IntVar(rps, "r", *rps, "alias for rps")

	threads := flag.Int("threads", 0, "number of concurrent threads for scenario")
	flag.IntVar(threads, "t", *threads, "alias for threads")

	duration := flag.Duration("duration", 0, "scenario duration, e.g. 5s")
	flag.DurationVar(duration, "d", *duration, "alias for duration")

	method := flag.String("method", "", "http method, e.g. POST")
	flag.StringVar(method, "m", *method, "alias for method")

	timeout := flag.Duration("timeout", 0, "http request timeout duration, e.g. 5s")

	flag.Parse()

	flags := &Flags{
		Debug:        *debug,
		Silent:       *silent,
		LogJSON:      *logJSON,
		SkipBanner:   *skipBanner,
		SkipExitCode: *skipExitCode,
		BrodAddr:     *brodAddr,
		Group:        *group,

		Output: *output,
		Format: *format,

		URL:      *url,
		RPS:      *rps,
		Threads:  *threads,
		Duration: *duration,
		Method:   *method,
		Timeout:  *timeout,

		Args: flag.Args(),
	}

	if flags.Debug {
		log.Debug().Any("flags", flags).Msg("flags parsed")
	}

	return flags
}
