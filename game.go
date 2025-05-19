package craps

// GameStats holds pass/craps/roll statistics for a single game trial.
type GameStats struct {
	ShooterCount uint
	RoundCount   uint
	RollCount    uint
	PassCount    uint
	CrapsCount   uint
}

// Game represents the state of a craps game trial.
type Game struct {
	roller   Roller
	point    uint
	lastRoll DiceRoll
	Stats    GameStats
}

func NewGame(r Roller) *Game {
	return &Game{roller: r}
}

func (g *Game) IsComeOut() bool {
	return g.point == 0
}

func (g *Game) rollDice() DiceRoll {
	roll := g.roller()
	g.lastRoll = roll
	g.Stats.RollCount++
	return roll
}

func (g *Game) reset() {
	g.point = 0
	g.lastRoll = DiceRoll{}
}

// Run executes the game until the player busts or the maximum rolls are reached.
// It loops through shooters until a terminal condition is met.
func (g *Game) Run(p *Player, maxRolls int) error {
	for {
		// come-out roll
		g.reset()
		g.Stats.ShooterCount++
		g.Stats.RoundCount++
		if err := p.strategy.PlaceBets(p, g); err != nil {
			return err
		}
		roll := g.rollDice()
		if roll.IsPoint() {
			g.point = roll.Value
			break
		} else {
			if roll.IsPass() {
				g.Stats.PassCount++
			}
			if roll.IsCraps() {
				g.Stats.CrapsCount++
			}
			p.settleBets(roll, g)
		}
		// trying to hit point
		for {
			if err := p.strategy.PlaceBets(p, g); err != nil {
				return err
			}
			roll := g.rollDice()
			if roll.Value == g.point {
				g.Stats.PassCount++
				break
			}
			if roll.Value == 7 {
				break
			}
			p.settleBets(roll, g)
		}

		if maxRolls > 0 && int(g.Stats.RollCount) >= maxRolls {
			break
		}
	}
	return nil
}
