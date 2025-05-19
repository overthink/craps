package craps

import (
	"log/slog"
)

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
	ID       int
	Bankroll float64
	strategy Strategy
	bets     []Bet
	Stats    PlayerStats
}

func (p *Player) LogValue() slog.Value {
	return slog.GroupValue(
		slog.Int("id", p.ID),
		slog.Float64("bankroll", p.Bankroll),
	)
}

// NewPlayer creates a new Player with the given id, bankroll, and strategy.
func NewPlayer(id int, bank float64, start Strategy) *Player {
	p := &Player{ID: id, Bankroll: bank, strategy: start}
	p.Stats.BankrollMax = bank
	p.Stats.BankrollMin = bank

	return p
}

// settleBets updates the player's bankroll and stats based on bet outcomes for
// the given roll and game state.
func (p *Player) settleBets(roll DiceRoll, g *Game) {
	var remaining []Bet
	betSettled := false
	for _, bet := range p.bets {
		bet.Update(roll, g)

		switch bet.Status() {
		case BetStatusWon:
			betSettled = true
			g.log.Info("bet won", "bet", bet)
			p.Bankroll += bet.Pays() + bet.Amount()
			p.Stats.WinCount++
		case BetStatusLost:
			betSettled = true
			g.log.Info("bet lost", "bet", bet)
			p.Stats.LossCount++
		default:
			g.log.Debug("bet not yet resolved", "bet", bet)
			remaining = append(remaining, bet)
		}
	}

	p.bets = remaining
	if p.Bankroll > p.Stats.BankrollMax {
		p.Stats.BankrollMax = p.Bankroll
	}

	if p.Bankroll < p.Stats.BankrollMin {
		p.Stats.BankrollMin = p.Bankroll
	}
	if betSettled {
		g.log.Info("bets settled", "player", p)
	}
}
