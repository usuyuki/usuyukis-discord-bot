package discord

import (
	"context"

	"github.com/bwmarrin/discordgo"

	"github.com/usuyuki/usuyukis-discord-bot/internal/interface/discordbot"
)

// AdminPermissionChecker はギルドメンバーが管理者権限を持つかどうかを判定する関数
type AdminPermissionChecker func(s *discordgo.Session, guildID, userID string) (bool, error)

// DefaultAdminPermissionChecker はdiscordgoのMember権限からAdministrator権限の有無を判定する
func DefaultAdminPermissionChecker(s *discordgo.Session, guildID, userID string) (bool, error) {
	perms, err := s.State.UserChannelPermissions(userID, guildID)
	if err != nil {
		perms, err = s.UserChannelPermissions(userID, guildID)
		if err != nil {
			return false, err
		}
	}
	return perms&discordgo.PermissionAdministrator != 0, nil
}

// RegisterEventBridge はdiscordgoのイベントをrouterへ変換して配送するハンドラをセッションに登録する。
// discordgo固有の型変換・State操作はこの関数に閉じ込め、interface/discordbot以下は
// discordgoを一切importしない
func RegisterEventBridge(s *discordgo.Session, router *discordbot.Router, checkAdmin AdminPermissionChecker) {
	s.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if m.Author == nil || m.Author.Bot {
			return
		}
		mentionsBot := false
		for _, u := range m.Mentions {
			if s.State.User != nil && u.ID == s.State.User.ID {
				mentionsBot = true
				break
			}
		}

		isAdmin := false
		if m.GuildID != "" {
			ok, err := checkAdmin(s, m.GuildID, m.Author.ID)
			if err == nil {
				isAdmin = ok
			}
		}

		botID := ""
		if s.State.User != nil {
			botID = s.State.User.ID
		}

		router.DispatchMessage(context.Background(), discordbot.IncomingMessage{
			GuildID:       m.GuildID,
			ChannelID:     m.ChannelID,
			AuthorID:      m.Author.ID,
			Content:       m.Content,
			MentionsBotID: mentionsBot,
			BotID:         botID,
			IsAdmin:       isAdmin,
		})
	})

	previousEmojis := map[string]map[string]bool{} // guildID -> emojiID set

	s.AddHandler(func(s *discordgo.Session, e *discordgo.GuildEmojisUpdate) {
		prev := previousEmojis[e.GuildID]
		current := make(map[string]bool, len(e.Emojis))
		var added []string
		for _, em := range e.Emojis {
			current[em.ID] = true
			if prev != nil && !prev[em.ID] {
				added = append(added, em.Name)
			}
		}
		previousEmojis[e.GuildID] = current

		if prev == nil || len(added) == 0 {
			return
		}
		router.DispatchEmojiUpdate(context.Background(), discordbot.IncomingEmojiUpdate{
			GuildID:         e.GuildID,
			AddedEmojiNames: added,
		})
	})
}
