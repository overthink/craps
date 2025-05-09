package main

import "errors"

// Odds represents odds for a bet. The numbers are net / independent of amount
// wagered.
type Odds struct {
	Win  uint
	Loss uint
}

func NewOdds(win, loss uint) (Odds, error) {
	if loss == 0 {
		return Odds{}, errors.New("loss must be non-zero")
	}
	return Odds{
		Win:  win,
		Loss: loss,
	}, nil
}

type Bet struct {
	Amount float64
	Pays   Odds
	// TODO: add true odds?
}

func NewBet(amount float64, win, loss uint) (Bet, error) {
	pays, err := NewOdds(win, loss)
	if err != nil {
		return Bet{}, err
	}
	return Bet{
		Amount: amount,
		Pays:   pays,
	}, nil
}

func (b Bet) Payout() float64 {
	return (float64(b.Pays.Win) / float64(b.Pays.Loss) * b.Amount) + b.Amount
}
