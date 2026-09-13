package cli

import "testing"

func TestFormatSeconds(t *testing.T) {
	cases := map[int]string{
		0:    "0s",
		59:   "59s",
		60:   "1min",
		3599: "59min",
		3600: "1h00min",
		3660: "1h01min",
		7325: "2h02min",
	}
	for seconds, want := range cases {
		if got := formatSeconds(seconds); got != want {
			t.Errorf("formatSeconds(%d) = %q, quer %q", seconds, got, want)
		}
	}
}
