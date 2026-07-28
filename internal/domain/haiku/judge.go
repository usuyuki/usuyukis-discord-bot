package haiku

// pattern は俳句の定型（5-7-5）モーラ数の区切り
var pattern = [3]int{5, 7, 5}

// Judge は形態素ごとのモーラ数の列が5-7-5（合計17拍）に区切れるかどうかを判定する。
// 各形態素（読みの単位）を跨いでモーラを分割することは不自然なため、
// 形態素の境界を尊重したまま、句切れが5拍・12拍(5+7)・17拍の位置とちょうど一致する
// 組み合わせが存在するかを探索する。
func Judge(moraCountsByWord []int) bool {
	total := 0
	for _, c := range moraCountsByWord {
		total += c
	}
	if total != pattern[0]+pattern[1]+pattern[2] {
		return false
	}

	boundaries := cumulativeBoundaries(moraCountsByWord)
	firstBreak := pattern[0]
	secondBreak := pattern[0] + pattern[1]
	return boundaries[firstBreak] && boundaries[secondBreak]
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
