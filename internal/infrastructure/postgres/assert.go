package postgres

import (
	emojiUC "github.com/usuyuki/usuyukis-discord-bot/internal/usecase/emoji"
	keywordUC "github.com/usuyuki/usuyukis-discord-bot/internal/usecase/keyword"
	notifychannelUC "github.com/usuyuki/usuyukis-discord-bot/internal/usecase/notifychannel"
	voicecallUC "github.com/usuyuki/usuyukis-discord-bot/internal/usecase/voicecall"
)

// コンパイル時にportを満たしていることを保証する
var (
	_ keywordUC.Repository            = (*KeywordRepository)(nil)
	_ notifychannelUC.Repository      = (*NotifyChannelRepository)(nil)
	_ emojiUC.NotifyChannelFinder     = (*NotifyChannelRepository)(nil)
	_ voicecallUC.NotifyChannelFinder = (*NotifyChannelRepository)(nil)
)
