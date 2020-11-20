package value

import (
	"fmt"
	"math"
	"strings"
)

func Round(v float64, round int) float64 {
	return math.Round(v*math.Pow10(round)) / math.Pow10(round)
}

func Convert(a, b Measure, v float64) float64 {
	return v * a.Float64() / b.Float64()
}

func ConvertRound(a, b Measure, v float64, round int) float64 {
	return Round(Convert(a, b, v), round)
}

func ConvertFromOne(m Measure, v float64) float64 {
	return Convert(1, m, v)
}

func ConvertFromOneRound(m Measure, v float64, round int) float64 {
	return ConvertRound(1, m, v, round)
}

func ConvertToOne(m Measure, v float64) float64 {
	return Convert(m, 1, v)
}

func ConvertToOneRound(m Measure, v float64, round int) float64 {
	return ConvertRound(m, 1, v, round)
}

func DigitsAfterDot(v float64) int {
	vs := strings.Split(fmt.Sprintf("%v", v), ".")
	if len(vs) == 1 {
		return 0
	}
	return len(vs[1])
}
