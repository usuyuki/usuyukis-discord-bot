package discord

import (
	"context"
	"strings"
	"sync"

	"github.com/bwmarrin/discordgo"

	"github.com/usuyuki/usuyukis-discord-bot/internal/domain/emoji"
	"github.com/usuyuki/usuyukis-discord-bot/internal/interface/discordbot"
)

// AdminPermissionChecker はギルドメンバーが管理者権限を持つかどうかを判定する関数
type AdminPermissionChecker func(s *discordgo.Session, guildID, userID string) (bool, error)

// detectsMentionsBot はmentionUserIDsに構造化メンションとしてbotIDが含まれるか、
// もしくはcontentの先頭単語が"@"を除いてbotNameに一致するかを判定する。
// 後者は過去メッセージのコピペ等で構造化メンションが失われ地の文の"@botName"表記に
// なってしまった場合の救済措置で、誤爆を避けるため先頭単語との一致のみ許容する
func detectsMentionsBot(mentionUserIDs []string, content, botID, botName string) bool {
	if botID != "" {
		for _, id := range mentionUserIDs {
			if id == botID {
				return true
			}
		}
	}
	if botName == "" {
		return false
	}
	fields := strings.Fields(content)
	return len(fields) > 0 && strings.EqualFold(strings.TrimPrefix(fields[0], "@"), botName)
}

// resolveGuild はState経由でギルド情報を取得し、State未キャッシュならREST APIへ
// フォールバックする。チャンネル制限・管理者判定など複数箇所で必要となる共通処理
func resolveGuild(s *discordgo.Session, guildID string) (*discordgo.Guild, error) {
	guild, err := s.State.Guild(guildID)
	if err != nil {
		return s.Guild(guildID)
	}
	return guild, nil
}

// DefaultAdminPermissionChecker はGuildのOwnerIDおよびMemberのロールから
// Administrator権限の有無を判定する。
// discordgo.Session.UserChannelPermissionsはチャンネル単位の実効権限を計算する
// ものでchannelIDを要求するため、guildIDを渡すと対象チャンネルが解決できずエラーになる。
// このチェックはチャンネルに依存しないギルド全体の管理者権限を見たいので、
// Guild/Member情報から直接権限ビットを集約する
func DefaultAdminPermissionChecker(s *discordgo.Session, guildID, userID string) (bool, error) {
	guild, err := resolveGuild(s, guildID)
	if err != nil {
		return false, err
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
		botID := ""
		botName := ""
		if s.State.User != nil {
			botID = s.State.User.ID
			botName = s.State.User.Username
		}

		mentions := make([]string, 0, len(m.Mentions))
		for _, u := range m.Mentions {
			mentions = append(mentions, u.ID)
		}
		mentionsBot := detectsMentionsBot(mentions, m.Content, botID, botName)

		isAdmin := false
		if m.GuildID != "" {
			ok, err := checkAdmin(s, m.GuildID, m.Author.ID)
			if err == nil {
				isAdmin = ok
			}
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

	// Bot自身が最初に付けたリアクション（提案メッセージへの初期✅）による無限ループ・
	// 誤カウントを避けるため、リアクションしたのがBot自身であるイベントは無視する
	s.AddHandler(func(s *discordgo.Session, e *discordgo.MessageReactionAdd) {
		if s.State.User != nil && e.UserID == s.State.User.ID {
			return
		}
		router.DispatchReactionAdd(context.Background(), discordbot.IncomingReactionAdd{
			ChannelID: e.ChannelID,
			MessageID: e.MessageID,
		})
	})

	// guildID -> emojiID set。プロセスローカルに前回状態を保持し、差分から追加された
	// 絵文字のみ通知する。Botがギルドから退出/キックされた際はGuildDeleteで該当エントリを
	// 削除するため、参加中のギルド数に比例したメモリ使用量に収まる
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

	// Botがギルドから退出/キックされた際、previousEmojisに残り続ける不要なエントリを削除する。
	// Unavailable=trueの場合はDiscord側の障害によるギルド一時利用不可であり、Botは脱退して
	// いないため削除しない（削除するとGuildEmojisUpdate再開時にprev==nilとなり、既存の
	// 全絵文字が追加されたという誤通知を招く）
	s.AddHandler(func(s *discordgo.Session, e *discordgo.GuildDelete) {
		if e.Unavailable {
			return
		}
		mu.Lock()
		delete(previousEmojis, e.ID)
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
