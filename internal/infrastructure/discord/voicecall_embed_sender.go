package discord

import (
	"context"
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/usuyuki/usuyukis-discord-bot/internal/domain/voicecall"
)

// voiceCallStartedColor/voiceCallEndedColorはEmbedの左帯色。開始と終了を色で
// 視覚的に区別できるよう、それぞれオレンジ・紫を割り当てる
const (
	voiceCallStartedColor = 0xE67E22
	voiceCallEndedColor   = 0x9B59B6
)

// jst は日本標準時（UTC+9）。実行環境にtzdataが無くても解決できるようFixedZoneで定義する
var jst = time.FixedZone("Asia/Tokyo", 9*60*60)

// formatJST はtをJSTの日時文字列に整形する
func formatJST(t time.Time) string {
	return t.In(jst).Format("2006-01-02 15:04:05")
}

// formatDuration はdをHH:MM:SS形式の文字列に整形する
func formatDuration(d time.Duration) string {
	total := int64(d.Seconds())
	h := total / 3600
	m := (total % 3600) / 60
	s := total % 60
	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}

// VoiceCallEmbedSender はusecase/voicecall.EmbedSender portのdiscordgo実装
type VoiceCallEmbedSender struct {
	session *discordgo.Session
}

// NewVoiceCallEmbedSender はVoiceCallEmbedSenderを生成する
func NewVoiceCallEmbedSender(session *discordgo.Session) *VoiceCallEmbedSender {
	return &VoiceCallEmbedSender{session: session}
}

// SendCallStarted は通話開始をEmbed形式で指定チャンネルへ送信する
func (s *VoiceCallEmbedSender) SendCallStarted(ctx context.Context, channelID string, notice voicecall.StartedNotice) error {
	embed := &discordgo.MessageEmbed{
		Title: "通話開始",
		Color: voiceCallStartedColor,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "チャンネル", Value: notice.VoiceChannelName, Inline: true},
			{Name: "始めた人", Value: notice.StarterName, Inline: true},
			{Name: "開始時間", Value: formatJST(notice.StartedAt), Inline: true},
		},
	}
	if notice.StarterAvatarURL != "" {
		embed.Thumbnail = &discordgo.MessageEmbedThumbnail{URL: notice.StarterAvatarURL}
	}
	if _, err := s.session.ChannelMessageSendEmbed(channelID, embed); err != nil {
		return fmt.Errorf("discord: failed to send voice call started embed: %w", err)
	}
	return nil
}

// SendCallEnded は通話終了をEmbed形式で指定チャンネルへ送信する
func (s *VoiceCallEmbedSender) SendCallEnded(ctx context.Context, channelID string, notice voicecall.EndedNotice) error {
	fields := []*discordgo.MessageEmbedField{
		{Name: "チャンネル", Value: notice.VoiceChannelName, Inline: true},
	}
	if notice.DurationKnown {
		fields = append(fields, &discordgo.MessageEmbedField{Name: "通話時間", Value: formatDuration(notice.Duration), Inline: true})
	}
	embed := &discordgo.MessageEmbed{
		Title:  "通話終了",
		Color:  voiceCallEndedColor,
		Fields: fields,
	}
	if _, err := s.session.ChannelMessageSendEmbed(channelID, embed); err != nil {
		return fmt.Errorf("discord: failed to send voice call ended embed: %w", err)
	}
	return nil
}
