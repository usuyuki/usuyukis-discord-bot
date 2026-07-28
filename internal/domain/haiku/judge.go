package haiku

// HaikuPattern は俳句の定型（5-7-5）モーラ数の区切り。
// usecase層はハードコードした定数を持たず、この値を参照することでdomain層との定義の重複・乖離を防ぐ
var HaikuPattern = []int{5, 7, 5}

// TankaPattern は短歌の定型（5-7-5-7-7）モーラ数の区切り。用途はHaikuPatternと同様
var TankaPattern = []int{5, 7, 5, 7, 7}

// Judge は形態素ごとのモーラ数の列が5-7-5（合計17拍）に区切れるかどうかを判定する。
// 各形態素（読みの単位）を跨いでモーラを分割することは不自然なため、
// 形態素の境界を尊重したまま、句切れが5拍・12拍(5+7)・17拍の位置とちょうど一致する
// 組み合わせが存在するかを探索する。
func Judge(moraCountsByWord []int) bool {
	return matchesPattern(moraCountsByWord, HaikuPattern)
}

// JudgeTanka は形態素ごとのモーラ数の列が5-7-5-7-7（合計31拍）に区切れるかどうかを判定する。
// Judgeと同様に形態素の境界を尊重して句切れ位置を探索する。
func JudgeTanka(moraCountsByWord []int) bool {
	return matchesPattern(moraCountsByWord, TankaPattern)
}

// matchesPattern はmoraCountsByWordの合計がpatternの合計拍数と一致し、
// かつpatternの各句切れ位置（累積拍数）がすべて形態素境界と一致するかを判定する
func matchesPattern(moraCountsByWord []int, pattern []int) bool {
	if len(pattern) == 0 {
		return false
	}
	total := 0
	for _, c := range moraCountsByWord {
		total += c
	}
	patternTotal := 0
	for _, p := range pattern {
		patternTotal += p
	}
	if total != patternTotal {
		return false
	}

	boundaries := cumulativeBoundaries(moraCountsByWord)
	cumulative := 0
	for _, p := range pattern[:len(pattern)-1] {
		cumulative += p
		if !boundaries[cumulative] {
			return false
		}
	}
	return true
}

// cumulativeBoundaries は形態素境界ごとの累積モーラ数の集合を真偽値マップで返す
func cumulativeBoundaries(moraCountsByWord []int) map[int]bool {
	boundaries := map[int]bool{0: true}
	sum := 0
	for _, c := range moraCountsByWord {
		sum += c
		boundaries[sum] = true
	}
	return boundaries
}

// Split はwordsをpattern（各句の拍数の列）の句切れ位置で分割し、句ごとに表層形を
// 連結した文字列のスライスを返す。wordsの形態素境界がpatternの句切れ位置と
// 一致しない場合はok=falseを返す
func Split(words []Word, pattern []int) (result []string, ok bool) {
	moraCounts := make([]int, len(words))
	for i, w := range words {
		moraCounts[i] = w.MoraCount
	}
	if !matchesPattern(moraCounts, pattern) {
		return nil, false
	}

	phrases := make([]string, len(pattern))
	wordIdx := 0
	cumulative := 0
	for i, p := range pattern {
		target := cumulative + p
		var b []byte
		for wordIdx < len(words) && cumulative < target {
			b = append(b, words[wordIdx].Surface...)
			cumulative += words[wordIdx].MoraCount
			wordIdx++
		}
		phrases[i] = string(b)
		cumulative = target
	}
	return phrases, true
}

// EvaluateBestSplit finds a partition of words into len(pattern) contiguous groups
// that minimizes the sum of absolute differences between the group mora counts and the pattern.
func EvaluateBestSplit(words []Word, pattern []int) (bestCounts []int) {
	if len(pattern) == 0 {
		return nil
	}
	if len(words) < len(pattern) {
		counts := make([]int, len(pattern))
		for _, w := range words {
			counts[0] += w.MoraCount
		}
		return counts
	}

	bestScore := -1
	bestCounts = make([]int, len(pattern))

	var search func(wordIdx int, groupIdx int, currentCounts []int)
	search = func(wordIdx int, groupIdx int, currentCounts []int) {
		if wordIdx == len(words) && groupIdx == len(pattern) {
			score := 0
			for i := range len(pattern) {
				diff := currentCounts[i] - pattern[i]
				if diff < 0 {
					score += -diff
				} else {
					score += diff
				}
			}
			if bestScore == -1 || score < bestScore {
				bestScore = score
				copy(bestCounts, currentCounts)
			}
			return
		}

		remainingGroups := len(pattern) - groupIdx
		remainingWords := len(words) - wordIdx
		if remainingWords < remainingGroups {
			return
		}
		if wordIdx == len(words) || groupIdx == len(pattern) {
			return
		}

		sum := 0
		maxK := remainingWords - remainingGroups + 1
		for k := 1; k <= maxK; k++ {
			sum += words[wordIdx+k-1].MoraCount
			currentCounts[groupIdx] = sum
			search(wordIdx+k, groupIdx+1, currentCounts)
		}
	}

	search(0, 0, make([]int, len(pattern)))
	return bestCounts
}
