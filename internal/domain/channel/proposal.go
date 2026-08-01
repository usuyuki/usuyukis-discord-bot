package channel

// IsApproved はapproverCount（提案メッセージに反応したユニークユーザー数、提案者自身を含む）が
// requiredApprovals以上かどうかを判定する
func IsApproved(approverCount, requiredApprovals int) bool {
	return approverCount >= requiredApprovals
}
