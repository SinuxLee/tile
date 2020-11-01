package decimal

import (
	"math/rand"
	"time"
)

// RandomByWeight 基于权重的随机
func RandomByWeight(weight []int) int {
	length := 0
	for _, num := range weight {
		length += num
	}

	randMake := rand.New(rand.NewSource(time.Now().UnixNano()))
	size := len(weight)
	for i := 0; i < size; i++ {
		randNum := randMake.Int() % length
		if randNum < weight[i] {
			return i
		}
		length -= weight[i]
	}

	return randMake.Int() % len(weight)
}
