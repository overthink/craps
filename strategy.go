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

	p.bets = append(p.bets, NewPassLineBet(5))
	p.Bankroll -= 5
	p.Stats.TotalWagered += 5
	p.Stats.BetCount++

	return nil
}
