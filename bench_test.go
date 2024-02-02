// SPDX-FileCopyrightText: Copyright 2024 Prasad Tengse
// SPDX-License-Identifier: MIT

package autotune_test

import (
	"math"
	"testing"
)

func primeNumbers(max int) []int {
	var primes []int

	for i := 2; i < max; i++ {
		isPrime := true

		for j := 2; j <= int(math.Sqrt(float64(i))); j++ {
			if i%j == 0 {
				isPrime = false
				break
			}
		}

		if isPrime {
			primes = append(primes, i)
		}
	}

	return primes
}

var num = 1000

func BenchmarkPrimeNumbers(b *testing.B) {
	for i := 0; i < b.N; i++ {
		primeNumbers(num)
	}
}
