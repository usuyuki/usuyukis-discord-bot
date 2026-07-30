package discord

import (
	"context"
	"sync"

	"github.com/bwmarrin/discordgo"

	"github.com/usuyuki/usuyukis-discord-bot/internal/domain/emoji"
	"github.com/usuyuki/usuyukis-discord-bot/internal/interface/discordbot"
)

// AdminPermissionChecker はギルドメンバーが管理者権限を持つかどうかを判定する関数
type AdminPermissionChecker func(s *discordgo.Session, guildID, userID string) (bool, error)

// DefaultAdminPermissionChecker はGuildのOwnerIDおよびMemberのロールから
// Administrator権限の有無を判定する。
// discordgo.Session.UserChannelPermissionsはチャンネル単位の実効権限を計算する
// ものでchannelIDを要求するため、guildIDを渡すと対象チャンネルが解決できずエラーになる。
// このチェックはチャンネルに依存しないギルド全体の管理者権限を見たいので、
// Guild/Member情報から直接権限ビットを集約する
func DefaultAdminPermissionChecker(s *discordgo.Session, guildID, userID string) (bool, error) {
	guild, err := s.State.Guild(guildID)
	if err != nil {
		guild, err = s.Guild(guildID)
		if err != nil {
			return false, err
		}
	}
	if guild.OwnerID == userID {
		return true, nil
	}

	member, err := s.State.Member(guildID, userID)
	if err != nil {
		member, err = s.GuildMember(guildID, userID)
		if err != nil {
			return false, err
		}
	}

	var perms int64
	for _, role := range guild.Roles {
		if role.ID == guild.ID {
			perms |= role.Permissions
			break
		}
	}
	for _, roleID := range member.Roles {
		for _, role := range guild.Roles {
			if role.ID == roleID {
				perms |= role.Permissions
				break
			}
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
		botName := ""
		if s.State.User != nil {
			botID = s.State.User.ID
			botName = s.State.User.Username
		}

		router.DispatchMessage(context.Background(), discordbot.IncomingMessage{
			GuildID:       m.GuildID,
			ChannelID:     m.ChannelID,
			AuthorID:      m.Author.ID,
			Content:       m.Content,
			MentionsBotID: mentionsBot,
			BotID:         botID,
			BotName:       botName,
			IsAdmin:       isAdmin,
		})
	})

	// guildID -> emojiID set。プロセスローカルに前回状態を保持し、差分から追加された
	// 絵文字のみ通知する。Bot参加ギルド数に比例してメモリを使用するが、通常運用の
	// ギルド数では問題にならない規模と判断している
	var mu sync.Mutex
	previousEmojis := map[string]map[string]bool{}

	// 起動時およびBot参加時に現在の絵文字リストを初期化する
	s.AddHandler(func(s *discordgo.Session, e *discordgo.GuildCreate) {
		current := make(map[string]bool, len(e.Emojis))
		for _, em := range e.Emojis {
			current[em.ID] = true
		}
		mu.Lock()
		previousEmojis[e.ID] = current
		mu.Unlock()
	})

	s.AddHandler(func(s *discordgo.Session, e *discordgo.GuildEmojisUpdate) {
		mu.Lock()
		prev := previousEmojis[e.GuildID]
		current := make(map[string]bool, len(e.Emojis))
		for _, em := range e.Emojis {
			current[em.ID] = true
		}

		var added []emoji.Emoji
		if prev != nil {
			var newEmojis []*discordgo.Emoji
			for _, em := range e.Emojis {
				if !prev[em.ID] {
					newEmojis = append(newEmojis, em)
				}
			}
			added = convertEmojis(newEmojis)
		}
		previousEmojis[e.GuildID] = current
		mu.Unlock()

		// prev == nil はこのギルドについてBot起動後はじめて受け取るイベント。
		// 「既存の全絵文字が追加された」という誤通知を避けるため、初回は必ず
		// スキップする（＝Bot再起動直後に発生した最初の絵文字追加は通知されない）
		// （GuildCreateで初期化されていれば、このルートには通常入らない）
		if prev == nil || len(added) == 0 {
			return
		}
		router.DispatchEmojiUpdate(context.Background(), discordbot.IncomingEmojiUpdate{
			GuildID:     e.GuildID,
			AddedEmojis: added,
		})
	})
}
