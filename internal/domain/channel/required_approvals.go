package channel

import "fmt"

// defaultRequiredApprovals はギルドが未設定の場合に使う必要承認人数のデフォルト値
const defaultRequiredApprovals = 2

// RequiredApprovals はチャンネル作成提案の可決に必要な承認人数（提案者自身を含む）を表す値オブジェクト
type RequiredApprovals struct {
	value int
}

// NewRequiredApprovals はrawを検証し、RequiredApprovalsを返す。1未満はエラーになる
func NewRequiredApprovals(raw int) (RequiredApprovals, error) {
	if raw < 1 {
		return RequiredApprovals{}, fmt.Errorf("channel: required approvals must be 1 or more, got %d", raw)
	}
	return RequiredApprovals{value: raw}, nil
}

// DefaultRequiredApprovals はギルドが必要承認人数を未設定の場合に使うデフォルト値を返す
func DefaultRequiredApprovals() RequiredApprovals {
	return RequiredApprovals{value: defaultRequiredApprovals}
}

// Int は必要承認人数の整数表現を返す
func (r RequiredApprovals) Int() int {
	return r.value
}
