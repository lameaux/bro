package app

import (
	"fmt"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/lameaux/bro/internal/client/config"
	"github.com/lameaux/bro/internal/client/runner"
	"github.com/lameaux/bro/internal/client/stats"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
)

func printResultsTable(conf *config.Config, results *stats.Stats, success bool) {
	fmt.Printf("Name: %s\nPath: %s\n", conf.Name, conf.FileName)

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Scenario", "Total", "Sent", "Success", "Failed", "Timeout", "Invalid", "Latency @P99", "Duration", "RPS", "Passed"})

	for _, scenario := range conf.Scenarios {
		counters := results.RequestCounters(scenario.Name)
		if counters == nil {
			log.Warn().
				Dict("scenario", zerolog.Dict().Str("name", scenario.Name)).
				Msg("missing stats")
			continue
		}

		t.AppendRow(table.Row{
			scenario.Name,
			counters.GetCounter(runner.CounterTotal),
			counters.GetCounter(runner.CounterSent),
			counters.GetCounter(runner.CounterSuccess),
			counters.GetCounter(runner.CounterFailed),
			counters.GetCounter(runner.CounterTimeout),
			counters.GetCounter(runner.CounterInvalid),
			fmt.Sprintf("%d ms", counters.GetLatencyAtPercentile(99)),
			counters.Duration,
			counters.Rps(),
			results.ThresholdsPassed(scenario.Name),
		})

	}

	t.SetStyle(table.StyleLight)
	t.Render()

	fmt.Printf("Total duration: %v\n", results.TotalDuration())

	if success {
		fmt.Println("OK")
	} else {
		fmt.Println("Failed")
	}
}
