package postgres

import (
	emojiUC "github.com/usuyuki/usuyukis-discord-bot/internal/usecase/emoji"
	haikuUC "github.com/usuyuki/usuyukis-discord-bot/internal/usecase/haiku"
	keywordUC "github.com/usuyuki/usuyukis-discord-bot/internal/usecase/keyword"
	notifychannelUC "github.com/usuyuki/usuyukis-discord-bot/internal/usecase/notifychannel"
)

// コンパイル時にportを満たしていることを保証する
var (
	_ keywordUC.Repository        = (*KeywordRepository)(nil)
	_ notifychannelUC.Repository  = (*NotifyChannelRepository)(nil)
	_ haikuUC.NotifyChannelFinder = (*NotifyChannelRepository)(nil)
	_ emojiUC.NotifyChannelFinder = (*NotifyChannelRepository)(nil)
)
