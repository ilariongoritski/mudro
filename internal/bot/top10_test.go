package bot

import "testing"

func TestIsTop10Line(t *testing.T) {
	cases := map[string]bool{
		"1. one":  true,
		"9. nine": true,
		"10. ten": true,
		"0. no":   false,
		"x. no":   false,
	}
	for in, want := range cases {
		if got := isTop10Line(in); got != want {
			t.Fatalf("isTop10Line(%q)=%v, want %v", in, got, want)
		}
	}
}
