package discord

import (
	"context"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/usuyuki/usuyukis-discord-bot/internal/domain/channel"
	"github.com/usuyuki/usuyukis-discord-bot/internal/domain/emoji"
	"github.com/usuyuki/usuyukis-discord-bot/internal/interface/discordbot"
)

// AdminPermissionChecker はギルドメンバーが管理者権限を持つかどうかを判定する関数
type AdminPermissionChecker func(s *discordgo.Session, guildID, userID string) (bool, error)

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

	// 一般ユーザーにもManageChannelsを持つロールを付与しチャンネル作成を許可する運用を前提とし、
	// 新規作成されたチャンネルが非公開であればチャンネル管理ロールのアクセスを剥奪する
	s.AddHandler(func(s *discordgo.Session, e *discordgo.ChannelCreate) {
		if e.GuildID == "" {
			return
		}
		router.DispatchChannelCreate(context.Background(), discordbot.IncomingChannelCreate{
			GuildID:   e.GuildID,
			ChannelID: e.ID,
			IsPrivate: channel.IsPrivate(e.GuildID, convertRoleOverwrites(e.PermissionOverwrites)),
			CreatorID: resolveChannelCreator(s, e.GuildID, e.ID),
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

// convertRoleOverwrites はdiscordgoのPermissionOverwriteのうちロール単位のものだけを
// domain層のchannel.Overwriteへ変換する。IsPrivate判定はロールへの拒否のみを見るため、
// メンバー単位のオーバーライドは対象外とする
func convertRoleOverwrites(overwrites []*discordgo.PermissionOverwrite) []channel.Overwrite {
	result := make([]channel.Overwrite, 0, len(overwrites))
	for _, ow := range overwrites {
		if ow.Type != discordgo.PermissionOverwriteTypeRole {
			continue
		}
		result = append(result, channel.Overwrite{RoleID: ow.ID, Deny: ow.Deny})
	}
	return result
}

// auditLogChannelCreateLimit は監査ログから取得するCHANNEL_CREATEエントリの件数。
// Discord APIの上限(100)に近い値を指定し、短時間に多数のチャンネルが連続作成されても
// 該当エントリが取得ウィンドウの外に押し出されにくくする
const auditLogChannelCreateLimit = 100

// auditLogRetryDelay は監査ログの反映ラグを吸収するための1回限りのリトライ待機時間
const auditLogRetryDelay = 500 * time.Millisecond

// resolveChannelCreator は監査ログから直近のCHANNEL_CREATEエントリを検索し、
// channelIDに一致するものがあれば作成者のユーザーIDを返す。監査ログの反映には若干の
// タイムラグがあるため、1回目で見つからなければ短い待機を挟んで1回だけ再試行する。
// 取得自体に失敗するケースもあり得るため、その場合は空文字を返す。
// 作成者が解決できない場合、呼び出し元はロール剥奪自体は行いつつ作成者への明示的な
// 許可は設定できないため、作成者が締め出されるリスクがある（利用可能な情報の範囲で
// できる限りラグを吸収した上での安全側フォールバック）
func resolveChannelCreator(s *discordgo.Session, guildID, channelID string) string {
	if userID := lookupChannelCreatorFromAuditLog(s, guildID, channelID); userID != "" {
		return userID
	}
	time.Sleep(auditLogRetryDelay)
	return lookupChannelCreatorFromAuditLog(s, guildID, channelID)
}

// lookupChannelCreatorFromAuditLog は監査ログを1回だけ問い合わせ、channelIDに
// 一致するCHANNEL_CREATEエントリのユーザーIDを返す。見つからない/取得失敗時は空文字を返す
func lookupChannelCreatorFromAuditLog(s *discordgo.Session, guildID, channelID string) string {
	auditLog, err := s.GuildAuditLog(guildID, "", "", int(discordgo.AuditLogActionChannelCreate), auditLogChannelCreateLimit)
	if err != nil {
		return ""
	}
	for _, entry := range auditLog.AuditLogEntries {
		if entry.TargetID == channelID {
			return entry.UserID
		}
	}
	return ""
}
