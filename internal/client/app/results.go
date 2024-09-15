package app

import (
	"fmt"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/lameaux/bro/internal/client/config"
	"github.com/lameaux/bro/internal/client/stats"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
)

func printResultsTable(conf *config.Config, results *stats.Stats, success bool) {
	fmt.Printf("Name: %s\nPath: %s\n", conf.Name, conf.FileName)

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Scenario", "Total", "Success", "Failed", "Timeout", "Invalid", "Latency @P99", "Duration", "RPS", "Passed"})

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
			counters.Counter(stats.CounterTotal),
			counters.Counter(stats.CounterSuccess),
			counters.Counter(stats.CounterFailed),
			counters.Counter(stats.CounterTimeout),
			counters.Counter(stats.CounterInvalid),
			fmt.Sprintf("%d ms", counters.LatencyAtPercentile(99)),
			results.Duration(scenario.Name),
			results.Rps(scenario.Name),
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
