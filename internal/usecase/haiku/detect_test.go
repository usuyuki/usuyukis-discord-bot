package haiku

import (
	"context"
	"testing"

	"github.com/usuyuki/usuyukis-discord-bot/internal/domain/haiku"
)

type fakeAnalyzer struct {
	words []haiku.Word
	err   error
}

func (f *fakeAnalyzer) AnalyzeWords(text string) ([]haiku.Word, error) {
	return f.words, f.err
}

func (f *fakeAnalyzer) Name() string {
	return "fake-analyzer"
}

type fakeSender struct {
	sentChannelID string
	sentContent   string
	called        bool
}

func (f *fakeSender) SendMessage(ctx context.Context, channelID, content string) error {
	f.sentChannelID = channelID
	f.sentContent = content
	f.called = true
	return nil
}

func TestIgnorePattern_ReplaceAllString(t *testing.T) {
	tests := []struct {
		name string
		body string
		want string
	}{
		{
			name: "正常系: 半角スペース区切りの本文はスペースが除去される",
			body: "古池や 蛙飛び込む 水の音",
			want: "古池や蛙飛び込む水の音",
		},
		{
			name: "正常系: 全角スペース区切りの本文は全角スペースが除去される",
			body: "古池や　蛙飛び込む　水の音",
			want: "古池や蛙飛び込む水の音",
		},
		{
			name: "正常系: 全角・半角スペースが混在していてもすべて除去される",
			body: "古池や　蛙飛び込む 水の音",
			want: "古池や蛙飛び込む水の音",
		},
		{
			name: "正常系: 句読点・記号もスペースとあわせて除去される",
			body: "古池や、蛙飛び込む！水の音。",
			want: "古池や蛙飛び込む水の音",
		},
		{
			name: "正常系: 全角の疑問符・カギ括弧・中黒などの記号も除去される",
			body: "「古池や」？蛙・飛び込む…水の音〜",
			want: "古池や蛙飛び込む水の音",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ignorePattern.ReplaceAllString(tt.body, "")
			if got != tt.want {
				t.Errorf("ignorePattern.ReplaceAllString() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestUseCase_Detect(t *testing.T) {
	haikuWords := []haiku.Word{
		{Surface: "古池", MoraCount: 3},
		{Surface: "や", MoraCount: 2},
		{Surface: "蛙", MoraCount: 3},
		{Surface: "飛び込む", MoraCount: 4},
		{Surface: "水の", MoraCount: 3},
		{Surface: "音", MoraCount: 2},
	}
	tankaWords := []haiku.Word{
		{Surface: "たら", MoraCount: 2},
		{Surface: "ちねの", MoraCount: 3},
		{Surface: "母が", MoraCount: 3},
		{Surface: "つりたる", MoraCount: 4},
		{Surface: "青蚊帳", MoraCount: 5},
		{Surface: "すがしい", MoraCount: 4},
		{Surface: "朝の", MoraCount: 3},
		{Surface: "かぜの", MoraCount: 3},
		{Surface: "中に", MoraCount: 2},
		{Surface: "ゐる", MoraCount: 2},
	}
	notWords := []haiku.Word{
		{Surface: "今日", MoraCount: 2},
		{Surface: "はいい", MoraCount: 3},
		{Surface: "天気", MoraCount: 3},
		{Surface: "ですね", MoraCount: 3},
	}

	tests := []struct {
		name              string
		words             []haiku.Word
		channelID         string
		messageBody       string
		wantDetected      bool
		wantSentChannelID string
		wantContent       string
	}{
		{
			name:              "正常系: 575判定でtrueなら投稿元チャンネルへ川柳として通知する",
			words:             haikuWords,
			channelID:         "test-channel",
			messageBody:       "古池や 蛙飛び込む 水の音",
			wantDetected:      true,
			wantSentChannelID: "test-channel",
			wantContent:       "川柳を検知しました:\n「古池や　蛙飛び込む　水の音」",
		},
		{
			name:              "正常系: 57577判定でtrueなら短歌として通知する",
			words:             tankaWords,
			channelID:         "test-channel",
			messageBody:       "たらちねの 母がつりたる 青蚊帳 すがしい朝の かぜの中にゐる",
			wantDetected:      true,
			wantSentChannelID: "test-channel",
			wantContent:       "短歌を検知しました:\n「たらちねの　母がつりたる　青蚊帳　すがしい朝の　かぜの中にゐる」",
		},
		{
			name:         "異常系: 575にも57577にもならなければ送信しない",
			words:        notWords,
			channelID:    "test-channel",
			messageBody:  "今日はいい天気ですね",
			wantDetected: false,
		},
		{
			name:              "正常系: --debugをつけるとデバッグ情報が出力される",
			words:             haikuWords,
			channelID:         "test-channel",
			messageBody:       "古池や 蛙飛び込む 水の音 --debug",
			wantDetected:      true,
			wantSentChannelID: "test-channel",
			wantContent:       "川柳を検知しました:\n「古池や　蛙飛び込む　水の音」\n\n【デバッグ: 使用形態素解析器】\nfake-analyzer\n\n【デバッグ: 形態素解析結果】\n```text\n古池\t\t\t(3拍)\nや\t\t\t(2拍)\n蛙\t\t\t(3拍)\n飛び込む\t\t\t(4拍)\n水の\t\t\t(3拍)\n音\t\t\t(2拍)\n```",
		},
		{
			name:              "異常系: --debugをつけると575でなくてもデバッグ情報が出力される",
			words:             notWords,
			channelID:         "test-channel",
			messageBody:       "今日はいい天気ですね --debug",
			wantDetected:      false,
			wantSentChannelID: "test-channel",
			wantContent:       "川柳・短歌として検知できませんでした。\n\n【デバッグ: 字余り・字足らず判定】\n川柳判定❌: 期待:5,7,5　結果:2,6,3\n短歌判定❌: 期待:5,7,5,7,7　結果:0,0,0,0,11\n\n【デバッグ: 使用形態素解析器】\nfake-analyzer\n\n【デバッグ: 形態素解析結果】\n```text\n今日\t\t\t(2拍)\nはいい\t\t\t(3拍)\n天気\t\t\t(3拍)\nですね\t\t\t(3拍)\n```",
		},
		{
			name:              "正常系: -debug（ハイフン1つ）をつけてもデバッグ情報が出力される（iPhoneで--が入力しにくいため）",
			words:             haikuWords,
			channelID:         "test-channel",
			messageBody:       "古池や 蛙飛び込む 水の音 -debug",
			wantDetected:      true,
			wantSentChannelID: "test-channel",
			wantContent:       "川柳を検知しました:\n「古池や　蛙飛び込む　水の音」\n\n【デバッグ: 使用形態素解析器】\nfake-analyzer\n\n【デバッグ: 形態素解析結果】\n```text\n古池\t\t\t(3拍)\nや\t\t\t(2拍)\n蛙\t\t\t(3拍)\n飛び込む\t\t\t(4拍)\n水の\t\t\t(3拍)\n音\t\t\t(2拍)\n```",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer := &fakeAnalyzer{words: tt.words}
			sender := &fakeSender{}
			u := New(analyzer, sender)

			detected, err := u.Detect(context.Background(), "g1", tt.channelID, tt.messageBody)
			if err != nil {
				t.Fatalf("Detect() unexpected error = %v", err)
			}
			if detected != tt.wantDetected {
				t.Fatalf("Detect() = %v, want %v", detected, tt.wantDetected)
			}
			if tt.wantContent != "" {
				if !sender.called {
					t.Fatal("Detect() expected SendMessage to be called, but it wasn't")
				}
				if sender.sentChannelID != tt.wantSentChannelID {
					t.Errorf("Detect() sentChannelID = %q, want %q", sender.sentChannelID, tt.wantSentChannelID)
				}
				if sender.sentContent != tt.wantContent {
					t.Errorf("Detect() sentContent = %q, want %q", sender.sentContent, tt.wantContent)
				}
			} else if sender.called {
				t.Error("Detect() expected SendMessage not to be called, but it was")
			}
		})
	}
}
