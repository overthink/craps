package main

import (
	"log/slog"
)

type ShooterStats struct {
	Rounds uint
	Rolls  uint
	Passes uint
	Craps  uint
}

// Shooter represents a player in craps that is rolling. It starts fresh on a
// come out roll, and ends when they seven out.
// TODO: we could also call this "Hand", thinking on it
type Shooter struct {
	ID    uint
	Point uint // 0 if unset
	Stats ShooterStats
	Roll  func() DiceRoll
	// lucky shooter info could be here too
}

func (s *Shooter) Run() {
	log := slog.With("shooterId", s.ID)
	log.Info("shooter start")
come_out:
	for {
		log.Info("come out roll")
		s.Stats.Rounds++
		roll := s.Roll()
		s.Stats.Rolls++
		log.Info("roll", "roll", roll)
		if roll.IsPoint() {
			log.Info("set point", "roll", roll)
			s.Point = roll.Value
			break
		}
		if roll.IsPass() {
			s.Stats.Passes++
		}
		if roll.IsCraps() {
			s.Stats.Craps++
		}
	}

	for {
		roll := s.Roll()
		s.Stats.Rolls++
		log.Info("roll", "roll", roll)
		if roll.Value == s.Point {
			log.Info("point hit", "roll", roll)
			s.Stats.Passes++
			s.Point = 0
			goto come_out
		}
		if roll.Value == 7 {
			log.Info("seven out")
			break
		}
	}
	log.Info("shooter done", "shooterStats", s.Stats)
}

// Odds represents odds for a bet. The numbers are net / independent of amount
// wagered.
type Odds struct {
	Win  uint
	Loss uint
}

type Bet struct {
	Wager uint
	Pays  Odds
	// TODO: add true odds?
}

type DiceRoll struct {
	Value uint
	Hard  bool
}

func (r DiceRoll) IsPoint() bool {
	return r.Value == 4 || r.Value == 5 || r.Value == 6 || r.Value == 8 || r.Value == 9 || r.Value == 10
}

func (r DiceRoll) IsPass() bool {
	return r.Value == 7 || r.Value == 11
}

func (r DiceRoll) IsCraps() bool {
	return r.Value == 2 || r.Value == 3 || r.Value == 12
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

func (s *Simulation) NewShooter() Shooter {
	s.ShooterCount++
	return Shooter{
		ID:   s.ShooterCount,
		Roll: s.Roll,
	}
}

func main() {
	sim := NewSimulation(9671111)
	for range 3 {
		shooter := sim.NewShooter()
		shooter.Run()
	}
	slog.Info("exiting", "shooterCount", sim.ShooterCount)
}
