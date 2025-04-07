package util

import (
	"math"
)

func CustomRound(solAmount float64) float64 {
	integerPart := int(solAmount)
	fraction := solAmount - float64(integerPart)

	candidates := []float64{0.0, 0.5, 1.0}

	closest := candidates[0]
	minDiff := math.Abs(fraction - closest)

	for _, c := range candidates {
		diff := math.Abs(fraction - c)
		if diff < minDiff {
			closest = c
			minDiff = diff
		}
	}

	return float64(integerPart) + closest
}
