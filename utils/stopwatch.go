package utils

import (
	"fmt"
	"time"
)

type Timing struct {
	Label   string
	Elapsed time.Duration
	Percent float64
}

type Stopwatch struct {
	startTimes map[string]time.Time
	timings    []Timing
	total      time.Duration
}

// NewStopwatch creates a new stopwatch
func NewStopwatch() *Stopwatch {
	return &Stopwatch{
		startTimes: make(map[string]time.Time),
	}
}

// Start begins timing for a label
func (sw *Stopwatch) Start(label string) {
	sw.startTimes[label] = time.Now()
}

// Stop ends timing for a label and stores the elapsed time
func (sw *Stopwatch) Stop(label string) {
	start, ok := sw.startTimes[label]
	if !ok {
		fmt.Printf("No start time recorded for label: %s\n", label)
		return
	}
	elapsed := time.Since(start)
	sw.total += elapsed
	sw.timings = append(sw.timings, Timing{
		Label:   label,
		Elapsed: elapsed,
	})
	delete(sw.startTimes, label)
}

// Finalize calculates percentages
func (sw *Stopwatch) Finalize() {
	for i := range sw.timings {
		sw.timings[i].Percent = (float64(sw.timings[i].Elapsed) / float64(sw.total)) * 100
	}
}

// PrintTable outputs the collected timings as a formatted table
func (sw *Stopwatch) PrintTable() {
	sw.Finalize()

	fmt.Println("┌──────────────────────────┬──────────────┬──────────────┐")
	fmt.Println("│ Label                    │ Elapsed Time │ Percentage   │")
	fmt.Println("├──────────────────────────┼──────────────┼──────────────┤")
	for _, t := range sw.timings {
		fmt.Printf("│ %-24s │ %-12s │ %6.2f %%     │\n", t.Label, t.Elapsed, t.Percent)
	}
	fmt.Println("└──────────────────────────┴──────────────┴──────────────┘")
}
