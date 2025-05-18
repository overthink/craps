package craps

import (
	"encoding/csv"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"sync"

	"log/slog"
)

// DEFAULT_ROLLS is the default maximum number of rolls per trial.
// A trial ends once this many rolls have been seen at the end of the current shooter.
const DEFAULT_ROLLS = 1000

type Strategy interface {
	PlaceBets(shooter *Shooter) error
}

type PassLineStrategy struct{}

func (s *PassLineStrategy) PlaceBets(shooter *Shooter) error {
	if shooter.Bankroll < 5 {
		return errors.New("not enough money")
	}
	if !shooter.IsComeOut() {
		return nil
	}
	shooter.Bets = append(shooter.Bets, NewPassLineBet(5))
	return nil
}

type ShooterStats struct {
	Rounds uint
	Rolls  uint
	Passes uint
	Craps  uint
}

// Shooter encapsulates all the game state for a single shooter. It starts fresh
// on a come out roll, and ends when they seven out.
type Shooter struct {
	ID       uint
	Log      *slog.Logger
	Strategy Strategy
	Bankroll float64
	Point    uint // 0 means not set
	Stats    ShooterStats
	Roller   Roller
	Bets     []Bet
}

func (s *Shooter) IsComeOut() bool {
	return s.Point == 0
}

func (s *Shooter) rollDice() DiceRoll {
	roll := s.Roller()
	s.Log.Info("roll", "roll", roll)
	s.Stats.Rolls++
	return roll
}

func (s *Shooter) Run() error {
	s.Log.Info("-------- shooter start")

	// make bets
	// roll
	// settle bets (with current roll and prev game state)
	// update game state
	// repeat

come_out:
	for {
		s.Log.Info("-- come out roll")
		s.Stats.Rounds++
		if err := s.Strategy.PlaceBets(s); err != nil {
			return err
		}
		roll := s.rollDice()
		if roll.IsPoint() {
			s.Point = roll.Value
			s.Log.Info("set point", "point", s.Point)
			break
		}
		if roll.IsPass() {
			s.Log.Info("natural/pass")
			s.Stats.Passes++
		}
		if roll.IsCraps() {
			s.Log.Info("craps")
			s.Stats.Craps++
		}
	}

	for {
		if err := s.Strategy.PlaceBets(s); err != nil {
			return err
		}
		roll := s.rollDice()
		if roll.Value == s.Point {
			s.Log.Info("point hit", "point", roll.Value)
			s.Stats.Passes++
			s.Point = 0
			goto come_out
		}
		if roll.Value == 7 {
			s.Log.Info("seven out")
			break
		}
	}
	s.Log.Info("shooter done", "bankroll", s.Bankroll, "shooterStats", s.Stats)
	return nil
}

// NewRoller returns a dice-roller function seeded from 'seed'.
// Each call to the returned Roller yields a reproducible pair of d6 rolls.
func NewRoller(seed int64) Roller {
	r := rand.New(rand.NewSource(seed))
	return func() DiceRoll {
		a := r.Intn(6) + 1
		b := r.Intn(6) + 1
		return DiceRoll{Value: uint(a + b), Hard: a == b}
	}
}

// Config holds the command-line options for the experiment.
type Config struct {
	Trials        int
	Bankroll      float64
	Seed          int64
	StrategyNames []string
	// Rolls is the maximum number of rolls per trial.
	Rolls int
	Out   string
	Quiet bool
}

func Run(cfg Config) error {
	if cfg.Trials > 1 || cfg.Quiet {
		slog.SetDefault(slog.New(slog.DiscardHandler))
	}

	names := cfg.StrategyNames
	strats := make([]Strategy, len(names))
	for i, name := range names {
		name = strings.TrimSpace(name)
		names[i] = name
		switch name {
		case "test":
			strats[i] = &PassLineStrategy{}
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

	type result struct {
		strategy string
		profit   float64
	}

	resultsCh := make(chan result, cfg.Trials*len(strats))
	var wg sync.WaitGroup

	for trialIdx := range cfg.Trials {
		wg.Add(1)
		trialSeed := cfg.Seed + int64(trialIdx)

		go func() {
			defer wg.Done()

			for idx, strat := range strats {
				roller := NewRoller(trialSeed)
				totalRolls := 0
				finalBank := cfg.Bankroll

				for {
					shooter := Shooter{
						ID:       uint(idx),
						Log:      slog.With("trial", trialIdx, "shooter", idx),
						Strategy: strat,
						Roller:   roller,
						Bankroll: finalBank,
					}
					if err := shooter.Run(); err != nil {
						finalBank = shooter.Bankroll
						break
					}
					totalRolls += int(shooter.Stats.Rolls)
					finalBank = shooter.Bankroll
					if totalRolls >= cfg.Rolls {
						break
					}
				}

				net := finalBank - cfg.Bankroll
				resultsCh <- result{strategy: names[idx], profit: net}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(resultsCh)
	}()

	for r := range resultsCh {
		if err := writer.Write([]string{r.strategy, fmt.Sprintf("%.2f", r.profit)}); err != nil {
			return fmt.Errorf("failed to write record: %w", err)
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return fmt.Errorf("error writing output: %w", err)
	}
	return nil
}
