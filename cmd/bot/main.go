package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/usuyuki/usuyukis-discord-bot/internal/config"
	discordInfra "github.com/usuyuki/usuyukis-discord-bot/internal/infrastructure/discord"
	"github.com/usuyuki/usuyukis-discord-bot/internal/infrastructure/morph"
	"github.com/usuyuki/usuyukis-discord-bot/internal/infrastructure/postgres"
	"github.com/usuyuki/usuyukis-discord-bot/internal/interface/admin"
	"github.com/usuyuki/usuyukis-discord-bot/internal/interface/discordbot"
	emojiUC "github.com/usuyuki/usuyukis-discord-bot/internal/usecase/emoji"
	haikuUC "github.com/usuyuki/usuyukis-discord-bot/internal/usecase/haiku"
	keywordUC "github.com/usuyuki/usuyukis-discord-bot/internal/usecase/keyword"
	notifychannelUC "github.com/usuyuki/usuyukis-discord-bot/internal/usecase/notifychannel"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("bot: fatal error: %v", err)
	}
}

func run() error {
	cfg, err := config.LoadFromEnv()
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	migrationsDir, err := filepath.Abs("migrations")
	if err != nil {
		return err
	}
	if err := postgres.Migrate(migrationsDir, cfg.DatabaseURL); err != nil {
		return err
	}

	pool, err := postgres.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	analyzer, err := morph.NewKagomeAnalyzer()
	if err != nil {
		return err
	}

	session, err := discordInfra.NewSession(cfg.DiscordBotToken)
	if err != nil {
		return err
	}

	keywordRepo := postgres.NewKeywordRepository(pool)
	notifyChannelRepo := postgres.NewNotifyChannelRepository(pool)
	messageSender := discordInfra.NewMessageSender(session)
	guildCache := discordInfra.NewGuildCache(session)

	keywordUseCase := keywordUC.New(keywordRepo)
	notifyChannelUseCase := notifychannelUC.New(notifyChannelRepo)
	haikuUseCase := haikuUC.New(analyzer, messageSender)
	emojiUseCase := emojiUC.New(notifyChannelRepo, messageSender)

	router := discordbot.NewRouter()
	if cfg.DevMode {
		router.SetDevChannelID(cfg.DevChannelID)
		log.Printf("bot: dev mode enabled, only channel %s will be handled", cfg.DevChannelID)
	}
	router.RegisterMessageHandler(discordbot.NewKeywordHandler(keywordUseCase, messageSender))
	router.RegisterMessageHandler(discordbot.NewHaikuHandler(haikuUseCase))
	router.RegisterMessageHandler(discordbot.NewHelpHandler(messageSender))
	router.RegisterEmojiUpdateHandler(discordbot.NewEmojiHandler(emojiUseCase))

	discordInfra.RegisterEventBridge(session, router, discordInfra.DefaultAdminPermissionChecker)

	if err := session.Open(); err != nil {
		return err
	}
	defer func() {
		if cerr := session.Close(); cerr != nil {
			log.Printf("bot: failed to close discord session: %v", cerr)
		}
	}()
	log.Println("bot: discord session opened")

	adminServer, err := admin.NewServer(
		admin.NewDiscordGuildDirectory(guildCache),
		keywordUseCase,
		notifyChannelUseCase,
	)
	if err != nil {
		return err
	}

	// 管理画面は認証機構を持たない。コンテナ内部では0.0.0.0以外にbindすると
	// Dockerのポートマッピングが機能しなくなるためここでは全interfaceにbindし、
	// docker-compose.yml側でホストの全interfaceへポート公開してLAN内アクセスを
	// 許可する運用を前提としている。インターネットへは公開しないこと
	httpServer := &http.Server{
		Addr:    ":" + cfg.AdminPort,
		Handler: adminServer.Handler(),
	}

	serveErrCh := make(chan error, 1)
	go func() {
		log.Printf("bot: admin server listening on %s", httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serveErrCh <- err
			return
		}
		serveErrCh <- nil
	}()

	select {
	case <-ctx.Done():
		log.Println("bot: shutdown signal received")
	case err := <-serveErrCh:
		if err != nil {
			return err
		}
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("bot: admin server shutdown error: %v", err)
	}

	return nil
}
