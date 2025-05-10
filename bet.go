package main

// Odds represents odds for a bet. The numbers are net / independent of amount
// wagered.
type Odds struct {
	Win  uint
	Loss uint
}

func NewOdds(win, loss uint) Odds {
	if loss == 0 {
		// It's annoying to propagate an error here, and it's a programmer error anyway
		panic("odds loss number must be positive")
	}
	return Odds{
		Win:  win,
		Loss: loss,
	}
}

type BetStatus int

const (
	BetStatusUnresolved BetStatus = iota
	BetStatusWon
	BetStatusLost
)

type BetType int

const (
	BetTypePassLine BetType = iota
	BetTypeCome
	BetTypeDontCome
	BetTypePassLineOdds
	BetTypeComeOdds
	BetTypeDontComeOdds
	BetTypePlace4
	BetTypePlace5
	BetTypePlace6
	BetTypePlace8
	BetTypePlace9
	BetTypePlace10
	BetTypePlace4Odds
	BetTypePlace5Odds
	BetTypePlace6Odds
	BetTypePlace8Odds
	BetTypePlace9Odds
	BetTypePlace10Odds
	BetTypeField
	BetTypeFieldOdds
	BetTypeHard4
	BetTypeHard6
	BetTypeHard8
	BetTypeHard10
	BetTypeHard4Odds
	BetTypeHard6Odds
	BetTypeHard8Odds
	BetTypeHard10Odds
	BetTypeAny7
	BetTypeAnyCraps
	BetTypeAny7Odds
	BetTypeAnyCrapsOdds
)

type Bet interface {
	Status() BetStatus

	// Return is the amount returned to the shooter if the bet is won, including
	// the original wager. e.g. if you bet 10 on 3:2 odds, you get 15 (winnings)
	// + 10 (original wager) == 25 back.
	Return() float64

	Update(roll DiceRoll, s *Shooter)
}

type BaseBet struct {
	amount float64
	odds   Odds
	status BetStatus
}

func (b *BaseBet) Status() BetStatus {
	return b.status
}

func (b *BaseBet) Return() float64 {
	return b.amount + (float64(b.odds.Win) / float64(b.odds.Loss) * b.amount)
}

type PassLineBet struct {
	BaseBet
	// We store the point vs relying on the game state so this type can be used for come bets as well
	point uint
}

func NewPassLineBet(amount float64) *PassLineBet {
	return &PassLineBet{
		BaseBet: BaseBet{
			amount: amount,
			odds:   NewOdds(1, 1),
		},
	}
}

func (pl *PassLineBet) Update(roll DiceRoll, s *Shooter) {
	// If we have a point, check if we hit it
	if pl.point > 0 {
		if roll.Value == pl.point {
			pl.status = BetStatusWon
		}
		if roll.Value == 7 {
			pl.status = BetStatusLost
		}
		return
	}
	// If we don't already have a point, check if we won, lost, or set a point
	if roll.IsPass() {
		pl.status = BetStatusWon
		return
	}
	if roll.IsCraps() {
		pl.status = BetStatusLost
		return
	}
	// If we got here we set the point
	pl.point = roll.Value
}
