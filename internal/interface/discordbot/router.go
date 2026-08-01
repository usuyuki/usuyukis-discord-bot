package discordbot

import (
	"context"
	"log"
)

// Router は登録された全ハンドラへイベントをブロードキャストする薄いディスパッチャ。
// 新機能追加時はRegisterMessageHandler/RegisterEmojiUpdateHandlerで1行追加するだけでよい
type Router struct {
	messageHandlers     []MessageHandler
	emojiUpdateHandlers []EmojiUpdateHandler
	reactionAddHandlers []ReactionAddHandler
	// devChannelID が空でない場合、DispatchMessageはこのチャンネル以外の
	// メッセージを全ハンドラへ配送しない（dev mode）
	devChannelID string
}

// NewRouter はRouterを生成する
func NewRouter() *Router {
	return &Router{}
}

// SetDevChannelID はdev modeを有効化し、以後DispatchMessageがchannelID以外の
// メッセージを配送しないようにする。空文字を渡すとdev modeは無効化される
func (r *Router) SetDevChannelID(channelID string) {
	r.devChannelID = channelID
}

// RegisterMessageHandler はMessageHandlerを登録する
func (r *Router) RegisterMessageHandler(h MessageHandler) {
	r.messageHandlers = append(r.messageHandlers, h)
}

// RegisterEmojiUpdateHandler はEmojiUpdateHandlerを登録する
func (r *Router) RegisterEmojiUpdateHandler(h EmojiUpdateHandler) {
	r.emojiUpdateHandlers = append(r.emojiUpdateHandlers, h)
}

// RegisterReactionAddHandler はReactionAddHandlerを登録する
func (r *Router) RegisterReactionAddHandler(h ReactionAddHandler) {
	r.reactionAddHandlers = append(r.reactionAddHandlers, h)
}

// DispatchMessage は登録済み全MessageHandlerへメッセージイベントを配送する。
// dev modeが有効な場合、devChannelID以外からのメッセージは配送せず無視する。
// 1つのハンドラのエラーが他のハンドラの実行を妨げないようログ出力のみ行い処理を続行する
func (r *Router) DispatchMessage(ctx context.Context, msg IncomingMessage) {
	if r.devChannelID != "" && msg.ChannelID != r.devChannelID {
		return
	}
	for _, h := range r.messageHandlers {
		if err := h.HandleMessage(ctx, msg); err != nil {
			log.Printf("discordbot: message handler error: %v", err)
		}
	}
}

// DispatchEmojiUpdate は登録済み全EmojiUpdateHandlerへ絵文字更新イベントを配送する。
// 絵文字更新はチャンネルを持たないギルド単位のイベントのためdevChannelIDとの比較はできないが、
// 「dev mode中は指定チャンネル以外への誤爆を防ぐ」という目的に合わせ、dev modeが有効な間は
// 絵文字通知そのものを配送しない
func (r *Router) DispatchEmojiUpdate(ctx context.Context, ev IncomingEmojiUpdate) {
	if r.devChannelID != "" {
		return
	}
	for _, h := range r.emojiUpdateHandlers {
		if err := h.HandleEmojiUpdate(ctx, ev); err != nil {
			log.Printf("discordbot: emoji update handler error: %v", err)
		}
	}
}

// DispatchReactionAdd は登録済み全ReactionAddHandlerへリアクション追加イベントを配送する。
// dev modeが有効な場合、devChannelID以外のチャンネルでのリアクションは配送せず無視する
func (r *Router) DispatchReactionAdd(ctx context.Context, ev IncomingReactionAdd) {
	if r.devChannelID != "" && ev.ChannelID != r.devChannelID {
		return
	}
	for _, h := range r.reactionAddHandlers {
		if err := h.HandleReactionAdd(ctx, ev); err != nil {
			log.Printf("discordbot: reaction add handler error: %v", err)
		}
	}
}
