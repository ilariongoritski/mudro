package casino

import (
	"fmt"
	"strconv"
)

var rouletteRed = map[int]struct{}{
	1: {}, 3: {}, 5: {}, 7: {}, 9: {},
	12: {}, 14: {}, 16: {}, 18: {},
	19: {}, 21: {}, 23: {}, 25: {}, 27: {},
	30: {}, 32: {}, 34: {}, 36: {},
}

func drawRouletteNumber(fairness *Fairness) int {
	return DrawIntWithFairness(fairness, 37)
}

func rouletteColor(number int) string {
	if number == 0 {
		return "green"
	}
	if _, ok := rouletteRed[number]; ok {
		return "red"
	}
	return "black"
}

func buildRouletteDisplaySequence(winningNumber int) []int {
	sequence := make([]int, 0, 24)
	for len(sequence) < 23 {
		sequence = append(sequence, drawRouletteNumber(nil))
	}
	sequence = append(sequence, winningNumber)
	return sequence
}

func buildRouletteResultSequence(displaySequence []int, winningNumber int) []int {
	if len(displaySequence) == 0 {
		return []int{winningNumber}
	}
	start := len(displaySequence) - 12
	if start < 0 {
		start = 0
	}
	out := append([]int(nil), displaySequence[start:]...)
	if out[len(out)-1] != winningNumber {
		out = append(out, winningNumber)
	}
	return out
}

func roulettePayout(bet RouletteBet, winningNumber int, winningColor string) (int64, bool) {
	switch bet.BetType {
	case "straight":
		number, err := parseRouletteNumber(bet.BetValue)
		if err != nil || number != winningNumber {
			return 0, false
		}
		return bet.Stake * 36, true
	case "red":
		if winningColor == "red" {
			return bet.Stake * 2, true
		}
	case "black":
		if winningColor == "black" {
			return bet.Stake * 2, true
		}
	case "green":
		if winningColor == "green" {
			return bet.Stake * 36, true
		}
	case "odd":
		if winningNumber != 0 && winningNumber%2 == 1 {
			return bet.Stake * 2, true
		}
	case "even":
		if winningNumber != 0 && winningNumber%2 == 0 {
			return bet.Stake * 2, true
		}
	case "low":
		if winningNumber >= 1 && winningNumber <= 18 {
			return bet.Stake * 2, true
		}
	case "high":
		if winningNumber >= 19 && winningNumber <= 36 {
			return bet.Stake * 2, true
		}
	case "dozen1":
		if winningNumber >= 1 && winningNumber <= 12 {
			return bet.Stake * 3, true
		}
	case "dozen2":
		if winningNumber >= 13 && winningNumber <= 24 {
			return bet.Stake * 3, true
		}
	case "dozen3":
		if winningNumber >= 25 && winningNumber <= 36 {
			return bet.Stake * 3, true
		}
	}
	return 0, false
}

func parseRouletteNumber(raw string) (int, error) {
	n, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("invalid roulette number %q", raw)
	}
	if n < 0 || n > 36 {
		return 0, fmt.Errorf("roulette number must be between 0 and 36")
	}
	return n, nil
}
