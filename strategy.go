package craps

type PassLineStrategy struct{}

func (s *PassLineStrategy) Name() string {
	return "passline"
}

func (s *PassLineStrategy) PlaceBets(g *Game, p *Player) {
	if p.Bankroll < 5 {
		return
	}

	if !g.IsComeOut() {
		return
	}

	bet := NewPassLineBet(5)
	p.Bets = append(p.Bets, bet)
	p.Bankroll -= 5
	p.Stats.TotalWagered += 5
	p.Stats.BetCount++
	g.log.Info("bet placed", "bet", bet, "player", p)
}
