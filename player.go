package craps

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
	Bankroll float64
	strategy Strategy
	bets     []Bet
	Stats    PlayerStats
}

// NewPlayer creates a new Player with the given id, bankroll, and strategy.
func NewPlayer(id uint, bank float64, strat Strategy) *Player {
	p := &Player{ID: id, Bankroll: bank, strategy: strat}
	p.Stats.BankrollMax = bank
	p.Stats.BankrollMin = bank
	return p
}

// settleBets updates the player's bankroll and stats based on bet outcomes for the given roll and game state.
func (p *Player) settleBets(roll DiceRoll, g *Game) {
	var remaining []Bet
	for _, bet := range p.bets {
		bet.Update(roll, g)
		switch bet.Status() {
		case BetStatusWon:
			p.Bankroll += bet.Return()
			p.Stats.WinCount++
		case BetStatusLost:
			p.Stats.LossCount++
		default:
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
}
