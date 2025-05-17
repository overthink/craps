package main

import (
	"errors"
	"log/slog"
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

func main() {
	sim := NewSimulation(9671111)
	for range 1 {
		strategy := &testStrategy{}
		shooter := sim.NewShooter(strategy, 440)
		for i := range MAX_ROUNDS {
			if err := shooter.Run(); err != nil {
				slog.Info("shooter error", "error", err)
				break
			}
			if i == MAX_ROUNDS-1 {
				// Arguably an error, but a very conservative strategy will hit this
				slog.Info("max rounds reached")
			}
		}
	}
	slog.Info("exiting", "shooterCount", sim.ShooterCount)
}
