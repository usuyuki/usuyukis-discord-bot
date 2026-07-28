package haiku

// smallKana は直前の仮名と結合して1モーラを構成する拗音・小書き文字
var smallKana = map[rune]bool{
	'ゃ': true, 'ゅ': true, 'ょ': true,
	'ァ': true, 'ィ': true, 'ゥ': true, 'ェ': true, 'ォ': true,
	'ャ': true, 'ュ': true, 'ョ': true,
	'ぁ': true, 'ぃ': true, 'ぅ': true, 'ぇ': true, 'ぉ': true,
}

// SplitMorae はひらがな・カタカナの読みをモーラ（拍）単位に分割する。
// 促音「ッ/っ」・撥音「ン/ん」・長音「ー」は独立した1モーラとしてカウントし、
// 拗音（ゃゅょ等の小書き文字）は直前の文字と結合して1モーラとする。
func SplitMorae(reading string) []string {
	runes := []rune(reading)
	morae := make([]string, 0, len(runes))
	for _, r := range runes {
		if smallKana[r] && len(morae) > 0 {
			morae[len(morae)-1] += string(r)
			continue
		}
		morae = append(morae, string(r))
	}
	return morae
}

// Word は形態素解析で得られる形態素1つ分の情報。
// Surface は元の表層形（漢字仮名混じりの原文表記）、MoraCount はその読みのモーラ数
// Reading は読み仮名、POS は品詞情報
type Word struct {
	Surface   string
	Reading   string
	POS       string
	MoraCount int
}
