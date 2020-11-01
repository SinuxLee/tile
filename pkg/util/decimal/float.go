package decimal

import "math"

// Round 截取指定的小数位
func Round(native float64, decimal int) float64 {
	power := math.Pow10(decimal)
	return math.Trunc((native+0.5/power)*power) / power
}
