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

	// IsWorking returns true if the bet is live. e.g. a pass line bet is always
	// working, but a place bet is (by default) only working if a point is set.
	// Only has meaning if the bet is unresolved.
	IsWorking() bool

	// CanBeRemoved returns true if the bet can be picked up from the table.
	// Some bets are locked in forever (e.g. pass line) but others can be
	// removed (e.g. place bets). Only has meaning if the bet is unresolved.
	CanBeRemoved() bool

	// Return is the amount returned to the shooter if the bet is won, including
	// the original wager. e.g. if you bet 10 on 3:2 odds, you get 15 (winnings)
	// + 10 (original wager) == 25 back.
	Return() float64

	// Amount returns the amount wagered on the bet.
	Amount() float64

	// Update gives the bet a chance to update its state based on the roll that
	// just happened and the game state as it existed before the roll.
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

func (b *BaseBet) IsWorking() bool {
	return true
}

func (b *BaseBet) CanBeRemoved() bool {
	return false
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

func (pl *PassLineBet) Update(roll DiceRoll, g *Game) {
	if g.point == 0 {
		// Come-out roll
		if roll.IsPass() {
			pl.status = BetStatusWon
			return
		}

		if roll.IsCraps() {
			pl.status = BetStatusLost
			return
		}
	}
	// Check for point
	if roll.Value == g.point {
		pl.status = BetStatusWon
	}

	if roll.Value == 7 {
		pl.status = BetStatusLost
	}

	return
}
