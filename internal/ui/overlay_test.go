package ui

import "testing"

func TestOverlayMove(t *testing.T) {
	o := &overlay{items: []string{"a", "b", "c"}, pick: func(int) {}}

	o.move(-5)
	if o.cursor != 0 {
		t.Fatalf("cursor = %d, quer 0", o.cursor)
	}
	o.move(1)
	o.move(1)
	o.move(5)
	if o.cursor != 2 {
		t.Fatalf("cursor = %d, quer 2", o.cursor)
	}
}

func TestProgressBar(t *testing.T) {
	if got := progressBar(0, 4); got != "[░░░░]" {
		t.Errorf("0%% = %q", got)
	}
	if got := progressBar(1, 4); got != "[████]" {
		t.Errorf("100%% = %q", got)
	}
	if got := progressBar(0.5, 4); got != "[██░░]" {
		t.Errorf("50%% = %q", got)
	}
}
