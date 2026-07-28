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

// EvaluateBestSplit はwordsをlen(pattern)個の連続したグループに分割する組み合わせのうち、
// 各グループのモーラ数合計とpatternの対応する句の拍数との差の絶対値の総和が最小になる
// 内訳（グループごとのモーラ数合計）を返す。字余り・字足らずでJudge/Splitが失敗した際に
// 「どこがどれだけずれているか」をデバッグ表示するために使う。
// 全分割の総当たりだと単語数に対して指数時間になり長文メッセージでBotの応答が止まりかねないため、
// 「i番目の単語までをj個の句に分けたときの最小スコア」をdp[i][j]として動的計画法で多項式時間に抑える
func EvaluateBestSplit(words []Word, pattern []int) (bestCounts []int) {
	if len(pattern) == 0 {
		return nil
	}

	n := len(words)
	k := len(pattern)

	// prefix[i] はwords[0:i]のモーラ数合計。区間[start,i)の合計をprefix[i]-prefix[start]で求める
	prefix := make([]int, n+1)
	for i, w := range words {
		prefix[i+1] = prefix[i] + w.MoraCount
	}

	const unreachable = -1
	dp := make([][]int, n+1)
	// split[i][j] はdp[i][j]を達成する最後の句の開始単語インデックス（復元用）
	split := make([][]int, n+1)
	for i := range dp {
		dp[i] = make([]int, k+1)
		split[i] = make([]int, k+1)
		for j := range dp[i] {
			dp[i][j] = unreachable
		}
	}
	dp[0][0] = 0

	for j := 1; j <= k; j++ {
		for i := 1; i <= n; i++ {
			for start := range i {
				if dp[start][j-1] == unreachable {
					continue
				}
				groupMora := prefix[i] - prefix[start]
				diff := groupMora - pattern[j-1]
				if diff < 0 {
					diff = -diff
				}
				score := dp[start][j-1] + diff
				if dp[i][j] == unreachable || score < dp[i][j] {
					dp[i][j] = score
					split[i][j] = start
				}
			}
		}
	}

	bestCounts = make([]int, k)
	i := n
	for j := k; j >= 1; j-- {
		start := split[i][j]
		bestCounts[j-1] = prefix[i] - prefix[start]
		i = start
	}
	return bestCounts
}
