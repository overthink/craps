package craps

import (
	"encoding/csv"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"runtime"
	"strings"

	"log/slog"

	"golang.org/x/sync/errgroup"
)

// DEFAULT_ROLLS is the default maximum number of rolls per trial. A trial ends
// once this many rolls have been seen at the end of the current shooter.
// Let's say 2 rolls per minute and a two hour session by default.
const DEFAULT_ROLLS = 2 * 60 * 2

// Strategy defines the betting logic for a player during a game.
type Strategy interface {
	PlaceBets(p *Player, g *Game) error
}

type PassLineStrategy struct{}

func (s *PassLineStrategy) PlaceBets(p *Player, g *Game) error {
	if p.Bankroll < 5 {
		return errors.New("not enough money")
	}
	if !g.IsComeOut() {
		return nil
	}
	p.bets = append(p.bets, NewPassLineBet(5))
	p.Bankroll -= 5
	p.Stats.TotalWagered += 5
	p.Stats.BetCount++
	return nil
}

func NewRoller(seed int64) Roller {
	r := rand.New(rand.NewSource(seed))
	return func() DiceRoll {
		a := r.Intn(6) + 1
		b := r.Intn(6) + 1
		return DiceRoll{Value: uint(a + b), Hard: a == b}
	}
}

type Config struct {
	Trials        int
	Bankroll      float64
	Seed          int64
	StrategyNames []string
	// Rolls is the maximum number of rolls per trial.
	Rolls   int
	Out     string
	Verbose bool
}

type result struct {
	strategy string
	profit   float64
}

func Run(cfg Config) error {
	if !cfg.Verbose {
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

	results := make([]result, cfg.Trials*len(strats))
	var eg errgroup.Group
	eg.SetLimit(runtime.GOMAXPROCS(0))

	for trialIdx := range cfg.Trials {
		eg.Go(func() error {
			trialSeed := cfg.Seed + int64(trialIdx)
			for idx, strat := range strats {
				log := slog.With("trial", trialIdx, "strategy", names[idx])
				roller := NewRoller(trialSeed)
				game := NewGame(log, roller)
				player := NewPlayer(uint(idx), cfg.Bankroll, strat)
				if err := game.Run(player, cfg.Rolls); err != nil {
					return fmt.Errorf("failed to run game: %w", err)
				}
				net := player.Bankroll - cfg.Bankroll
				resultIdx := trialIdx*len(strats) + idx
				results[resultIdx] = result{strategy: names[idx], profit: net}
			}
			return nil
		})
	}

	err := eg.Wait()
	if err != nil {
		return fmt.Errorf("error running trials: %w", err)
	}
	if err := writeResults(cfg, results); err != nil {
		return fmt.Errorf("failed to write results: %w", err)
	}
	return nil
}

func writeResults(config Config, results []result) error {
	var writer *csv.Writer
	var f *os.File
	if config.Out != "" {
		var err error
		f, err = os.Create(config.Out)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer f.Close()
		writer = csv.NewWriter(f)
	} else {
		writer = csv.NewWriter(os.Stdout)
	}

	if err := writer.Write([]string{"strategy", "net_profit"}); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}
	for _, result := range results {
		if err := writer.Write([]string{result.strategy, fmt.Sprintf("%.2f", result.profit)}); err != nil {
			return fmt.Errorf("failed to write record: %w", err)
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return fmt.Errorf("error writing output: %w", err)
	}
	return nil
}
