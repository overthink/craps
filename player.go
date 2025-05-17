package main

// PlayerStats holds per-player win/loss statistics.
type PlayerStats struct {
	WinCount     uint
	LossCount    uint
	TotalWagered float64
	BetCount     uint
	BankrollMax  float64
	BankrollMin  float64
}

// Player represents the person betting at the table.
type Player struct {
	ID       uint
	bankroll float64
	strategy Strategy
	bets     []Bet
	stats    PlayerStats
}

// NewPlayer creates a new Player with the given id, bankroll, and strategy.
func NewPlayer(id uint, bank float64, strat Strategy) *Player {
	return &Player{ID: id, bankroll: bank, strategy: strat}
}
