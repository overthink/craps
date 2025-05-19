package craps

import (
	"testing"

	"github.com/stretchr/testify/require"
)

const delta = 0.005

func TestBaseBet(t *testing.T) {
	t.Parallel()

	tests := []struct {
		amount  float64
		win     uint
		loss    uint
		pays    float64
		isError bool
	}{
		{
			amount: 10,
			win:    3,
			loss:   2,
			pays:   15,
		},
		{
			amount: 10,
			win:    2,
			loss:   3,
			pays:   6.67,
		},
		{
			amount: 0,
			win:    3,
			loss:   2,
			pays:   0.0,
		},
		{
			amount: 12,
			win:    5,
			loss:   3,
			pays:   20,
		},
		{
			amount: 12.50,
			win:    5,
			loss:   3,
			pays:   20.83,
		},
		{
			amount:  1,
			win:     3,
			loss:    0,
			isError: true,
		},
		{
			amount: 1,
			win:    0,
			loss:   1,
			pays:   0,
		},
		{
			amount: 0,
			win:    2,
			loss:   1,
			pays:   0,
		},
	}
	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			t.Parallel()

			if tt.isError {
				require.Panics(t, func() {
					_ = BaseBet{
						amount: tt.amount,
						odds:   NewOdds(tt.win, tt.loss),
					}
				})

				return
			}

			b := BaseBet{
				amount: tt.amount,
				odds:   NewOdds(tt.win, tt.loss),
			}
			require.InDelta(t, tt.pays, b.Pays(), delta)
		})
	}
}
