package main

type DiceRoll struct {
	Value uint
	Hard  bool
}

type Roller func() DiceRoll

func (r DiceRoll) IsPoint() bool {
	return r.Value == 4 || r.Value == 5 || r.Value == 6 || r.Value == 8 || r.Value == 9 || r.Value == 10
}

func (r DiceRoll) IsPass() bool {
	return r.Value == 7 || r.Value == 11
}

func (r DiceRoll) IsCraps() bool {
	return r.Value == 2 || r.Value == 3 || r.Value == 12
}
