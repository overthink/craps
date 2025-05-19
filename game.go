package craps

import (
	"errors"
	"log/slog"
)

// GameStats holds pass/craps/roll statistics for a single game trial.
type GameStats struct {
	RoundCount uint
	RollCount  uint
	PassCount  uint
	CrapsCount uint
}

// Game represents the state of a craps game trial.
type Game struct {
	log    *slog.Logger
	roller Roller
	point  uint
	Stats  GameStats
}

func NewGame(log *slog.Logger, r Roller) *Game {
	return &Game{
		log:    log,
		roller: r,
	}
}

func (g *Game) IsComeOut() bool {
	return g.point == 0
}

func (g *Game) rollDice() DiceRoll {
	roll := g.roller()
	g.log.Info("rolled", "value", roll.Value, "hard", roll.Hard)

	g.Stats.RollCount++

	return roll
}

func (g *Game) reset() {
	g.point = 0
}

// Run executes the game until the player busts or the maximum rolls are reached.
// It loops through shooters until a terminal condition is met.
func (g *Game) Run(player *Player, maxRolls int) error {
ComeOutLoop:
	for {
		for _, bet := range player.bets {
			// TODO: real case? Do we have to take back non-working bets here?
			g.log.Info("unsettled working bets", "bet", bet)
			return errors.New("unsettled bets found at game start")
		}
		if maxRolls > 0 && g.Stats.RollCount >= uint(maxRolls) {
			g.log.Info("max rolls reached", "rolls", g.Stats.RollCount, "player", player)
			break ComeOutLoop
		}

		// come-out roll
		g.log.Info("---- come out", "player", player)
		g.reset()
		g.Stats.RoundCount++
		if err := player.strategy.PlaceBets(player, g); err != nil {
			return err
		}
		roll := g.rollDice()
		player.settleBets(roll, g)
		if roll.IsPoint() {
			g.log.Info("point set", "point", roll.Value)
			g.point = roll.Value
		} else {
			if roll.IsPass() {
				g.log.Info("pass")
				g.Stats.PassCount++
			}
			if roll.IsCraps() {
				g.log.Info("craps")
				g.Stats.CrapsCount++
			}
			player.settleBets(roll, g)
			continue ComeOutLoop
		}
		// trying to hit point
	PointLoop:
		for {
			if err := player.strategy.PlaceBets(player, g); err != nil {
				return err
			}
			roll := g.rollDice()
			player.settleBets(roll, g)
			if roll.Value == g.point {
				g.log.Info("point hit")
				g.Stats.PassCount++
				break PointLoop
			}
			if roll.Value == 7 {
				g.log.Info("seven out")
				break PointLoop
			}
		}
	}

	return nil
}
