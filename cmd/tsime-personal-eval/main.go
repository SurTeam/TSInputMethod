package main

import (
	"fmt"
	"os"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "tsime-personal-eval: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	samples := deterministicFixtures()
	fmt.Printf("tsime-personal-eval fixtures=%d (deterministic; not production accuracy)\n", len(samples))
	modes := []struct {
		name    ModeName
		useSGD  bool
		useFreq bool
	}{
		{ModeFast, false, false},
		{ModeUserFreq, false, true},
		{ModeSGD, true, false},
		{ModeUserFreqSGD, true, true},
	}
	for _, m := range modes {
		res := runMode(m.name, samples, m.useSGD, m.useFreq)
		fmt.Println(formatResult(res))
	}
	fmt.Println("note: predict→observe→train per sample; fixture-only metrics")
	return nil
}
