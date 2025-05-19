package craps

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
	BetTypeBuy4
	BetTypeBuy5
	BetTypePlace6
	BetTypePlace8
	BetTypeBuy9
	BetTypeBuy10
)

type Bet interface {
	Status() BetStatus
	Type() BetType

	// Return is the amount returned to the shooter if the bet is won, including
	// the original wager. e.g. if you bet 10 on 3:2 odds, you get 15 (winnings)
	// + 10 (original wager) == 25 back.
	Return() float64

	// Amount returns the amount wagered on the bet.
	Amount() float64
	// Update processes the given roll against the game state.
	Update(roll DiceRoll, g *Game)
}

type BaseBet struct {
	amount  float64
	odds    Odds
	status  BetStatus
	betType BetType
}

func (b *BaseBet) Status() BetStatus {
	return b.status
}

func (b *BaseBet) Type() BetType {
	return b.betType
}

func (b *BaseBet) Return() float64 {
	return b.amount + (float64(b.odds.Win) / float64(b.odds.Loss) * b.amount)
}

// Amount returns the wagered amount of the base bet.
func (b *BaseBet) Amount() float64 {
	return b.amount
}

type PassLineBet struct {
	BaseBet
	// We store the point vs relying on the game state so this type can be used for come bets as well
	point uint
}

func NewPassLineBet(amount float64) *PassLineBet {
	return &PassLineBet{
		BaseBet: BaseBet{
			amount:  amount,
			odds:    NewOdds(1, 1),
			betType: BetTypePassLine,
		},
	}
}

func (pl *PassLineBet) Update(roll DiceRoll, _ *Game) {
	// If the bet has its own point, check win/loss conditions
	if pl.point > 0 {
		if roll.Value == pl.point {
			pl.status = BetStatusWon
		}
		if roll.Value == 7 {
			pl.status = BetStatusLost
		}
		return
	}
	// Come-out roll: check pass or craps outcomes
	if roll.IsPass() {
		pl.status = BetStatusWon
		return
	}
	if roll.IsCraps() {
		pl.status = BetStatusLost
		return
	}
	// Otherwise set point for the bet
	pl.point = roll.Value
}
