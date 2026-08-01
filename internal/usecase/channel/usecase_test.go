package channel

import (
	"context"
	"errors"
	"testing"
)

// fakeCreator はテスト用のCreatorフェイク実装
type fakeCreator struct {
	called     bool
	gotGuildID string
	gotName    string
	err        error
}

func (f *fakeCreator) CreateTextChannel(ctx context.Context, guildID, name string) error {
	f.called = true
	f.gotGuildID = guildID
	f.gotName = name
	return f.err
}

// fakeMessenger はテスト用のProposalMessengerフェイク実装
type fakeMessenger struct {
	called       bool
	gotChannelID string
	gotContent   string
	returnMsgID  string
	err          error
}

func (f *fakeMessenger) SendProposal(ctx context.Context, channelID, content string) (string, error) {
	f.called = true
	f.gotChannelID = channelID
	f.gotContent = content
	return f.returnMsgID, f.err
}

// fakeCounter はテスト用のApprovalCounterフェイク実装
type fakeCounter struct {
	count int
	err   error
}

func (f *fakeCounter) CountUniqueReactors(ctx context.Context, channelID, messageID string) (int, error) {
	return f.count, f.err
}

// fakeProposalRepo はテスト用のProposalRepositoryフェイク実装（インメモリ）。
// TryResolveは実際のpostgres実装同様、resolvedがまだfalseの場合にのみtrueへ遷移させ
// claimed=trueを返すことで、二重解決防止のロジックをフェイク上でも再現する
type fakeProposalRepo struct {
	saved        []Proposal
	resolved     map[string]bool
	findResult   Proposal
	findFound    bool
	findErr      error
	saveErr      error
	resolveErr   error
	unresolveErr error
	// alreadyResolvedOnClaim はTryResolve呼び出し時点で強制的にclaimedをfalseにする
	// （並行して他の呼び出しが先に解決済みにした状況を再現するためのテスト用フラグ）
	alreadyResolvedOnClaim bool
}

func newFakeProposalRepo() *fakeProposalRepo {
	return &fakeProposalRepo{resolved: map[string]bool{}}
}

func (f *fakeProposalRepo) Save(ctx context.Context, p Proposal) error {
	if f.saveErr != nil {
		return f.saveErr
	}
	f.saved = append(f.saved, p)
	return nil
}

func (f *fakeProposalRepo) FindByMessage(ctx context.Context, channelID, messageID string) (Proposal, bool, error) {
	if f.findErr != nil {
		return Proposal{}, false, f.findErr
	}
	return f.findResult, f.findFound, nil
}

func (f *fakeProposalRepo) TryResolve(ctx context.Context, channelID, messageID string) (bool, error) {
	if f.resolveErr != nil {
		return false, f.resolveErr
	}
	key := channelID + "/" + messageID
	if f.alreadyResolvedOnClaim || f.resolved[key] {
		return false, nil
	}
	f.resolved[key] = true
	return true, nil
}

func (f *fakeProposalRepo) Unresolve(ctx context.Context, channelID, messageID string) error {
	if f.unresolveErr != nil {
		return f.unresolveErr
	}
	f.resolved[channelID+"/"+messageID] = false
	return nil
}

// fakeSettingRepo はテスト用のSettingRepositoryフェイク実装
type fakeSettingRepo struct {
	required    int
	found       bool
	err         error
	setCalled   bool
	setGuildID  string
	setRequired int
	setErr      error
}

func (f *fakeSettingRepo) Get(ctx context.Context, guildID string) (int, bool, error) {
	return f.required, f.found, f.err
}

func (f *fakeSettingRepo) Set(ctx context.Context, guildID string, requiredApprovals int) error {
	f.setCalled = true
	f.setGuildID = guildID
	f.setRequired = requiredApprovals
	return f.setErr
}

func newTestUseCase(creator Creator, messenger ProposalMessenger, counter ApprovalCounter, proposals ProposalRepository, settings SettingRepository) *UseCase {
	return New(creator, messenger, counter, proposals, settings)
}

func TestUseCase_Propose(t *testing.T) {
	t.Run("正常系: 妥当な名前なら提案メッセージを送信し保存する", func(t *testing.T) {
		creator := &fakeCreator{}
		messenger := &fakeMessenger{returnMsgID: "msg1"}
		repo := newFakeProposalRepo()
		u := newTestUseCase(creator, messenger, &fakeCounter{}, repo, &fakeSettingRepo{})

		if err := u.Propose(context.Background(), "g1", "c1", "user1", "general-chat"); err != nil {
			t.Fatalf("Propose() unexpected error = %v", err)
		}
		if !messenger.called {
			t.Fatal("Propose() should send a proposal message")
		}
		if len(repo.saved) != 1 {
			t.Fatalf("Propose() should save exactly 1 proposal, got %d", len(repo.saved))
		}
		got := repo.saved[0]
		if got.GuildID != "g1" || got.ChannelID != "c1" || got.MessageID != "msg1" || got.ChannelName != "general-chat" || got.ProposerID != "user1" {
			t.Errorf("saved proposal = %+v, unexpected", got)
		}
		if creator.called {
			t.Error("Propose() should not create the channel immediately")
		}
	})

	t.Run("異常系: 不正な名前を入れるとNewNameがエラーになり提案は送信されない", func(t *testing.T) {
		messenger := &fakeMessenger{}
		repo := newFakeProposalRepo()
		u := newTestUseCase(&fakeCreator{}, messenger, &fakeCounter{}, repo, &fakeSettingRepo{})

		err := u.Propose(context.Background(), "g1", "c1", "user1", "Invalid Name!")
		if err == nil {
			t.Fatal("Propose() expected error for invalid name, got nil")
		}
		if messenger.called {
			t.Error("Propose() should not send a message when name is invalid")
		}
	})
}

func TestUseCase_RecordReaction(t *testing.T) {
	t.Run("正常系: 承認数が必要数未満ならチャンネルは作成されない", func(t *testing.T) {
		creator := &fakeCreator{}
		repo := newFakeProposalRepo()
		repo.findResult = Proposal{GuildID: "g1", ChannelID: "c1", MessageID: "msg1", ChannelName: "general-chat", ProposerID: "user1"}
		repo.findFound = true
		counter := &fakeCounter{count: 1}
		settings := &fakeSettingRepo{required: 2, found: true}
		u := newTestUseCase(creator, &fakeMessenger{}, counter, repo, settings)

		if err := u.RecordReaction(context.Background(), "c1", "msg1"); err != nil {
			t.Fatalf("RecordReaction() unexpected error = %v", err)
		}
		if creator.called {
			t.Error("RecordReaction() should not create the channel when approvals are below the threshold")
		}
	})

	t.Run("正常系: 承認数が必要数に達するとチャンネルを作成し提案を解決済みにする", func(t *testing.T) {
		creator := &fakeCreator{}
		repo := newFakeProposalRepo()
		repo.findResult = Proposal{GuildID: "g1", ChannelID: "c1", MessageID: "msg1", ChannelName: "general-chat", ProposerID: "user1"}
		repo.findFound = true
		counter := &fakeCounter{count: 2}
		settings := &fakeSettingRepo{required: 2, found: true}
		u := newTestUseCase(creator, &fakeMessenger{}, counter, repo, settings)

		if err := u.RecordReaction(context.Background(), "c1", "msg1"); err != nil {
			t.Fatalf("RecordReaction() unexpected error = %v", err)
		}
		if !creator.called {
			t.Fatal("RecordReaction() should create the channel when approvals reach the threshold")
		}
		if creator.gotGuildID != "g1" || creator.gotName != "general-chat" {
			t.Errorf("Creator received guildID=%q name=%q, want g1/general-chat", creator.gotGuildID, creator.gotName)
		}
		if !repo.resolved["c1/msg1"] {
			t.Error("RecordReaction() should mark the proposal as resolved")
		}
	})

	t.Run("正常系: 必要承認人数がギルド未設定ならデフォルト値(2)が使われる", func(t *testing.T) {
		creator := &fakeCreator{}
		repo := newFakeProposalRepo()
		repo.findResult = Proposal{GuildID: "g1", ChannelID: "c1", MessageID: "msg1", ChannelName: "general-chat", ProposerID: "user1"}
		repo.findFound = true
		counter := &fakeCounter{count: 2}
		settings := &fakeSettingRepo{found: false}
		u := newTestUseCase(creator, &fakeMessenger{}, counter, repo, settings)

		if err := u.RecordReaction(context.Background(), "c1", "msg1"); err != nil {
			t.Fatalf("RecordReaction() unexpected error = %v", err)
		}
		if !creator.called {
			t.Fatal("RecordReaction() should create the channel using the default threshold")
		}
	})

	t.Run("異常系: 対象の提案が見つからなければ何もしない", func(t *testing.T) {
		creator := &fakeCreator{}
		repo := newFakeProposalRepo()
		repo.findFound = false
		u := newTestUseCase(creator, &fakeMessenger{}, &fakeCounter{count: 5}, repo, &fakeSettingRepo{required: 2, found: true})

		if err := u.RecordReaction(context.Background(), "c1", "msg1"); err != nil {
			t.Fatalf("RecordReaction() unexpected error = %v", err)
		}
		if creator.called {
			t.Error("RecordReaction() should not create the channel when the proposal is not found")
		}
	})

	t.Run("異常系: すでに解決済みの提案には再度チャンネル作成しない（二重作成防止）", func(t *testing.T) {
		creator := &fakeCreator{}
		repo := newFakeProposalRepo()
		repo.findResult = Proposal{GuildID: "g1", ChannelID: "c1", MessageID: "msg1", ChannelName: "general-chat", ProposerID: "user1", Resolved: true}
		repo.findFound = true
		counter := &fakeCounter{count: 5}
		u := newTestUseCase(creator, &fakeMessenger{}, counter, repo, &fakeSettingRepo{required: 2, found: true})

		if err := u.RecordReaction(context.Background(), "c1", "msg1"); err != nil {
			t.Fatalf("RecordReaction() unexpected error = %v", err)
		}
		if creator.called {
			t.Error("RecordReaction() should not create the channel twice for an already resolved proposal")
		}
	})

	t.Run("異常系: Creatorがエラーを返すとRecordReactionもエラーを返し、提案は未解決に戻る", func(t *testing.T) {
		wantErr := errors.New("boom")
		creator := &fakeCreator{err: wantErr}
		repo := newFakeProposalRepo()
		repo.findResult = Proposal{GuildID: "g1", ChannelID: "c1", MessageID: "msg1", ChannelName: "general-chat", ProposerID: "user1"}
		repo.findFound = true
		counter := &fakeCounter{count: 2}
		u := newTestUseCase(creator, &fakeMessenger{}, counter, repo, &fakeSettingRepo{required: 2, found: true})

		err := u.RecordReaction(context.Background(), "c1", "msg1")
		if !errors.Is(err, wantErr) {
			t.Fatalf("RecordReaction() error = %v, want %v", err, wantErr)
		}
		if repo.resolved["c1/msg1"] {
			t.Error("RecordReaction() should unresolve the proposal when channel creation fails, so a later reaction can retry")
		}
	})

	t.Run("異常系: 閾値到達と同時に他の呼び出しが先に解決権を得ていた場合はチャンネルを作成しない（二重作成防止）", func(t *testing.T) {
		creator := &fakeCreator{}
		repo := newFakeProposalRepo()
		repo.findResult = Proposal{GuildID: "g1", ChannelID: "c1", MessageID: "msg1", ChannelName: "general-chat", ProposerID: "user1"}
		repo.findFound = true
		repo.alreadyResolvedOnClaim = true
		counter := &fakeCounter{count: 2}
		u := newTestUseCase(creator, &fakeMessenger{}, counter, repo, &fakeSettingRepo{required: 2, found: true})

		if err := u.RecordReaction(context.Background(), "c1", "msg1"); err != nil {
			t.Fatalf("RecordReaction() unexpected error = %v", err)
		}
		if creator.called {
			t.Error("RecordReaction() should not create the channel when TryResolve loses the race (claimed=false)")
		}
	})
}

func TestUseCase_GetRequiredApprovals(t *testing.T) {
	t.Run("正常系: 設定済みならその値を返す", func(t *testing.T) {
		settings := &fakeSettingRepo{required: 3, found: true}
		u := newTestUseCase(&fakeCreator{}, &fakeMessenger{}, &fakeCounter{}, newFakeProposalRepo(), settings)

		got, err := u.GetRequiredApprovals(context.Background(), "g1")
		if err != nil {
			t.Fatalf("GetRequiredApprovals() unexpected error = %v", err)
		}
		if got != 3 {
			t.Errorf("GetRequiredApprovals() = %d, want 3", got)
		}
	})

	t.Run("異常系: 未設定ならデフォルト値(2)を返す", func(t *testing.T) {
		settings := &fakeSettingRepo{found: false}
		u := newTestUseCase(&fakeCreator{}, &fakeMessenger{}, &fakeCounter{}, newFakeProposalRepo(), settings)

		got, err := u.GetRequiredApprovals(context.Background(), "g1")
		if err != nil {
			t.Fatalf("GetRequiredApprovals() unexpected error = %v", err)
		}
		if got != 2 {
			t.Errorf("GetRequiredApprovals() = %d, want 2", got)
		}
	})
}

func TestUseCase_SetRequiredApprovals(t *testing.T) {
	t.Run("正常系: 1以上の値ならRepositoryへ保存する", func(t *testing.T) {
		settings := &fakeSettingRepo{}
		u := newTestUseCase(&fakeCreator{}, &fakeMessenger{}, &fakeCounter{}, newFakeProposalRepo(), settings)

		if err := u.SetRequiredApprovals(context.Background(), "g1", 3); err != nil {
			t.Fatalf("SetRequiredApprovals() unexpected error = %v", err)
		}
		if !settings.setCalled || settings.setGuildID != "g1" || settings.setRequired != 3 {
			t.Errorf("Repository received guildID=%q required=%d, want g1/3", settings.setGuildID, settings.setRequired)
		}
	})

	t.Run("異常系: 0以下を入れるとNewRequiredApprovalsがエラーになりRepositoryは呼ばれない", func(t *testing.T) {
		settings := &fakeSettingRepo{}
		u := newTestUseCase(&fakeCreator{}, &fakeMessenger{}, &fakeCounter{}, newFakeProposalRepo(), settings)

		err := u.SetRequiredApprovals(context.Background(), "g1", 0)
		if err == nil {
			t.Fatal("SetRequiredApprovals() expected error for non-positive value, got nil")
		}
		if settings.setCalled {
			t.Error("SetRequiredApprovals() should not call Repository when the value is invalid")
		}
	})
}
