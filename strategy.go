package craps

import "errors"

type PassLineStrategy struct{}

func (s *PassLineStrategy) Name() string {
	return "passline"
}

func (s *PassLineStrategy) PlaceBets(p *Player, g *Game) error {
	if p.Bankroll < 5 {
		return errors.New("not enough money")
	}

	if !g.IsComeOut() {
		return nil
	}

	bet := NewPassLineBet(5)
	p.bets = append(p.bets, bet)
	p.Bankroll -= 5
	p.Stats.TotalWagered += 5
	p.Stats.BetCount++
	g.log.Info("bet placed", "bet", bet, "player", p)
	return nil
}
