package exchange

import "testing"

func Test_Profitability(t *testing.T) {
	h, l := Pair{"", "", 1000}, Pair{"", "", 900}
	fee := 0.1
	conv := 0.5

	prof := Profitability(&h, &l, fee, conv)
	want := 0.405
	if prof != want {
		t.Errorf("got %f, want %f", prof, want)
	}
}
