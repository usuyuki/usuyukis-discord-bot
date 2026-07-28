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
}

// NewRouter はRouterを生成する
func NewRouter() *Router {
	return &Router{}
}

// RegisterMessageHandler はMessageHandlerを登録する
func (r *Router) RegisterMessageHandler(h MessageHandler) {
	r.messageHandlers = append(r.messageHandlers, h)
}

// RegisterEmojiUpdateHandler はEmojiUpdateHandlerを登録する
func (r *Router) RegisterEmojiUpdateHandler(h EmojiUpdateHandler) {
	r.emojiUpdateHandlers = append(r.emojiUpdateHandlers, h)
}

// DispatchMessage は登録済み全MessageHandlerへメッセージイベントを配送する。
// 1つのハンドラのエラーが他のハンドラの実行を妨げないようログ出力のみ行い処理を続行する
func (r *Router) DispatchMessage(ctx context.Context, msg IncomingMessage) {
	for _, h := range r.messageHandlers {
		if err := h.HandleMessage(ctx, msg); err != nil {
			log.Printf("discordbot: message handler error: %v", err)
		}
	}
}

// DispatchEmojiUpdate は登録済み全EmojiUpdateHandlerへ絵文字更新イベントを配送する
func (r *Router) DispatchEmojiUpdate(ctx context.Context, ev IncomingEmojiUpdate) {
	for _, h := range r.emojiUpdateHandlers {
		if err := h.HandleEmojiUpdate(ctx, ev); err != nil {
			log.Printf("discordbot: emoji update handler error: %v", err)
		}
	}
}
