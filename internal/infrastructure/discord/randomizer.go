package discord

import "math/rand"

// Randomizer はslot usecaseのRandomizer portをmath/rand経由で実装する
type Randomizer struct{}

// NewRandomizer はRandomizerを生成する
func NewRandomizer() *Randomizer {
	return &Randomizer{}
}

// Intn は[0, n)の範囲の乱数を1つ返す
func (r *Randomizer) Intn(n int) int {
	return rand.Intn(n)
}
