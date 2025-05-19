package main

import (
	"fmt"
	"os"

	"github.com/overthink/craps"
	"github.com/spf13/cobra"
)

func main() {
	var cfg craps.Config
	cmd := &cobra.Command{
		Use:   "craps",
		Short: "Run craps experiments",
		RunE: func(cmd *cobra.Command, args []string) error {
			return craps.Run(cfg)
		},
	}
	cmd.Flags().IntVar(&cfg.Trials, "trials", 1, "number of trials to run for each strategy")
	cmd.Flags().Float64Var(&cfg.Bankroll, "bankroll", 440, "starting bankroll for shooters")
	cmd.Flags().Int64Var(&cfg.Seed, "seed", 9671111, "base seed; trial seeds will be seed+trial")
	cmd.Flags().StringSliceVar(&cfg.StrategyNames, "strategies", []string{"passline"}, "comma-separated list of strategies to test")
	cmd.Flags().IntVar(&cfg.Rolls, "rolls", craps.DEFAULT_ROLLS, "maximum number of rolls per trial (trial stops after this many rolls once the shooter sevens out)")
	cmd.Flags().StringVar(&cfg.Out, "out", "", "output CSV file path (default stdout)")
	cmd.Flags().BoolVar(&cfg.Verbose, "verbose", false, "enable logging output")

	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
