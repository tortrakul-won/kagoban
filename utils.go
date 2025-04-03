package main

import (
	"crypto/rand"
	"encoding/hex"
)

func mapSlice(input []int, transform func(int) int) []int {
	result := make([]int, len(input))
	for i, v := range input {
		result[i] = transform(v)
	}
	return result
}

func randomHex(n int) string {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return ""
	}
	return "#" + hex.EncodeToString(bytes)
}

func clamp(minVal, val, maxVal int) int {
	return max(min(maxVal, val), minVal)
}
