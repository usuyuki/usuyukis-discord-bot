package discord

import (
	emojiUC "github.com/usuyuki/usuyukis-discord-bot/internal/usecase/emoji"
	haikuUC "github.com/usuyuki/usuyukis-discord-bot/internal/usecase/haiku"
)

// コンパイル時にportを満たしていることを保証する
var (
	_ haikuUC.MessageSender = (*MessageSender)(nil)
	_ emojiUC.MessageSender = (*MessageSender)(nil)
)
