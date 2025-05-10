package main

import (
	"log/slog"
)

type Strategy struct {
	//
}

func (s *Strategy) Bet(shooter *Shooter) {
	//
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
	point    uint // 0 if unset
	stats    ShooterStats
	roller   Roller
	bets     []Bet
}

func (s *Shooter) rollDice() DiceRoll {
	roll := s.roller()
	s.log.Info("roll", "roll", roll)
	s.stats.Rolls++
	return roll
}

func (s *Shooter) Run() {
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
		s.strategy.Bet(s)
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
		s.strategy.Bet(s)
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
	s.log.Info("shooter done", "shooterStats", s.stats)
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
	for range 3 {
		strategy := Strategy{}
		shooter := sim.NewShooter(strategy, 440)
		shooter.Run()
	}
	slog.Info("exiting", "shooterCount", sim.ShooterCount)
}
