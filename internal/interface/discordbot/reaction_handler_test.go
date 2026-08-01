package discordbot

import (
	"context"
	"errors"
	"testing"
)

// fakeChannelReactionUseCase はテスト用のUseCaseフェイク実装
type fakeChannelReactionUseCase struct {
	called       bool
	gotChannelID string
	gotMessageID string
	err          error
}

func (f *fakeChannelReactionUseCase) RecordReaction(ctx context.Context, channelID, messageID string) error {
	f.called = true
	f.gotChannelID = channelID
	f.gotMessageID = messageID
	return f.err
}

func TestReactionHandler_HandleReactionAdd(t *testing.T) {
	t.Run("正常系: リアクション追加イベントをUseCaseへそのまま委譲する", func(t *testing.T) {
		uc := &fakeChannelReactionUseCase{}
		h := NewReactionHandler(uc)

		ev := IncomingReactionAdd{ChannelID: "c1", MessageID: "msg1"}
		if err := h.HandleReactionAdd(context.Background(), ev); err != nil {
			t.Fatalf("HandleReactionAdd() unexpected error = %v", err)
		}
		if !uc.called || uc.gotChannelID != "c1" || uc.gotMessageID != "msg1" {
			t.Errorf("UseCase received channelID=%q messageID=%q, want c1/msg1", uc.gotChannelID, uc.gotMessageID)
		}
	})

	t.Run("異常系: UseCaseのエラーがそのまま返る", func(t *testing.T) {
		wantErr := errors.New("boom")
		uc := &fakeChannelReactionUseCase{err: wantErr}
		h := NewReactionHandler(uc)

		err := h.HandleReactionAdd(context.Background(), IncomingReactionAdd{ChannelID: "c1", MessageID: "msg1"})
		if !errors.Is(err, wantErr) {
			t.Fatalf("HandleReactionAdd() error = %v, want %v", err, wantErr)
		}
	})
}
