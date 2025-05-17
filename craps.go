package main

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"log/slog"

	"github.com/spf13/cobra"
)

const MAX_ROUNDS = 1000

type Strategy interface {
	PlaceBets(shooter *Shooter) error
}

type testStrategy struct{}

func (s *testStrategy) PlaceBets(shooter *Shooter) error {
	if shooter.bankroll < 5 {
		return errors.New("not enough money")
	}
	if !shooter.IsComeOut() {
		return nil
	}
	shooter.bets = append(shooter.bets, NewPassLineBet(5))
	return nil
}

type ShooterStats struct {
	Rounds uint
	Rolls  uint
	Passes uint
	Craps  uint
}

// Shooter encapsulates all the game state for a single shooter. It starts fresh
// on a come out roll, and ends when they seven out.  You could probably also
// call this "Hand" but it is more understandable to me as Shooter.
type Shooter struct {
	ID       uint
	log      *slog.Logger
	strategy Strategy
	bankroll float64
	point    uint // 0 means not set
	stats    ShooterStats
	roller   Roller
	bets     []Bet
}

func (s *Shooter) IsComeOut() bool {
	return s.point == 0
}

func (s *Shooter) rollDice() DiceRoll {
	roll := s.roller()
	s.log.Info("roll", "roll", roll)
	s.stats.Rolls++
	return roll
}

func (s *Shooter) Run() error {
	s.log.Info("-------- shooter start")

	// make bets
	// roll
	// settle bets (with current roll and prev game state)
	// update game state
	// repeat

come_out:
	for {
		s.log.Info("-- come out roll")
		s.stats.Rounds++
		if err := s.strategy.PlaceBets(s); err != nil {
			return err
		}
		roll := s.rollDice()
		if roll.IsPoint() {
			s.point = roll.Value
			s.log.Info("set point", "point", s.point)
			break
		}
		if roll.IsPass() {
			s.log.Info("natural/pass")
			s.stats.Passes++
		}
		if roll.IsCraps() {
			s.log.Info("craps")
			s.stats.Craps++
		}
	}

	for {
		if err := s.strategy.PlaceBets(s); err != nil {
			return err
		}
		roll := s.rollDice()
		if roll.Value == s.point {
			s.log.Info("point hit", "point", roll.Value)
			s.stats.Passes++
			s.point = 0
			goto come_out
		}
		if roll.Value == 7 {
			s.log.Info("seven out")
			break
		}
	}
	s.log.Info("shooter done", "bankroll", s.bankroll, "shooterStats", s.stats)
	return nil
}

type Simulation struct {
	ShooterCount uint
	prng         PRNG
}

func NewSimulation(seed int64) *Simulation {
	return &Simulation{
		prng: NewPRNG(seed),
	}
}

func (s *Simulation) Roll() DiceRoll {
	a, b := s.prng.Roll2()
	return DiceRoll{
		Value: uint(a + b),
		Hard:  a == b,
	}
}

func (s *Simulation) NewShooter(strategy Strategy, bankroll float64) Shooter {
	s.ShooterCount++
	return Shooter{
		ID:       s.ShooterCount,
		log:      slog.With("shooterId", s.ShooterCount),
		strategy: strategy,
		roller:   s.Roll,
		bankroll: bankroll,
	}
}

// Config holds the command-line options for the experiment.
type Config struct {
	Trials        int
	Bankroll      float64
	Seed          int64
	StrategyNames []string
	Out           string
	Quiet         bool
}

func run(cfg Config) error {
	if cfg.Trials > 1 || cfg.Quiet {
		slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{})))
	}

	names := cfg.StrategyNames
	strats := make([]Strategy, len(names))
	for i, name := range names {
		name = strings.TrimSpace(name)
		names[i] = name
		switch name {
		case "test":
			strats[i] = &testStrategy{}
		default:
			return fmt.Errorf("unknown strategy: %s", name)
		}
	}

	var writer *csv.Writer
	var f *os.File
	if cfg.Out != "" {
		var err error
		f, err = os.Create(cfg.Out)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer func() { _ = f.Close() }()
		writer = csv.NewWriter(f)
	} else {
		writer = csv.NewWriter(os.Stdout)
	}

	if err := writer.Write([]string{"strategy", "net_profit"}); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	for t := range cfg.Trials {
		for idx, strat := range strats {
			sim := NewSimulation(cfg.Seed + int64(t))
			shooter := sim.NewShooter(strat, cfg.Bankroll)
			for range MAX_ROUNDS {
				if err := shooter.Run(); err != nil {
					break
				}
			}
			net := shooter.bankroll - cfg.Bankroll
			if err := writer.Write([]string{names[idx], fmt.Sprintf("%.2f", net)}); err != nil {
				return fmt.Errorf("failed to write record: %w", err)
			}
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return fmt.Errorf("error writing output: %w", err)
	}
	return nil
}

func main() {
	var cfg Config
	cmd := &cobra.Command{
		Use:   "craps",
		Short: "Run craps experiments",
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cfg)
		},
	}
	cmd.Flags().IntVar(&cfg.Trials, "trials", 1, "number of trials to run for each strategy")
	cmd.Flags().Float64Var(&cfg.Bankroll, "bankroll", 440, "starting bankroll for shooters")
	cmd.Flags().Int64Var(&cfg.Seed, "seed", 9671111, "base seed; trial seeds will be seed+trial")
	cmd.Flags().StringSliceVar(&cfg.StrategyNames, "strategies", []string{"test"}, "comma-separated list of strategies to test")
	cmd.Flags().StringVar(&cfg.Out, "out", "", "output CSV file path (default stdout)")
	cmd.Flags().BoolVar(&cfg.Quiet, "quiet", false, "suppress logging output")

	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
