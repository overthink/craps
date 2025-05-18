package craps

// TableStats holds pass/craps/roll statistics at the table level.
type TableStats struct {
	ShooterCount uint
	RoundCount   uint
	RollCount    uint
	PassCount    uint
	CrapsCount   uint
}

// Table represents the craps table state.
type Table struct {
	roller   Roller
	point    uint
	lastRoll DiceRoll
	stats    TableStats
}

// NewTable creates a new Game with the provided roller.
func NewTable(r Roller) *Table {
	return &Table{roller: r}
}
