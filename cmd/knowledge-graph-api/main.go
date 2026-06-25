package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"knowledgegraph/internal/api"
	"knowledgegraph/internal/config"
	"knowledgegraph/internal/storage"
)

func main() {
	// 1. 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config load failed: %v", err)
	}
	log.Printf("starting knowledge-graph-service: %s", cfg)

	// 2. 初始化 MySQL（降级策略：DBRequired=false 时无库也能启动）
	var store *storage.Store
	if cfg.Database.DSN() != "" {
		candidate, err := storage.NewMySQLStore(cfg.Database.DSN())
		if err != nil {
			log.Printf("mysql: connect failed: %v", err)
		} else {
			// Ping
			pingCtx, pingCancel := context.WithTimeout(context.Background(), 5*time.Second)
			if err := candidate.Ping(pingCtx); err != nil {
				log.Printf("mysql: ping failed: %v", err)
				_ = candidate.Close()
				candidate = nil
			} else {
				pingCancel()
				// EnsureSchema
				schemaCtx, schemaCancel := context.WithTimeout(context.Background(), 10*time.Second)
				if err := candidate.EnsureSchema(schemaCtx); err != nil {
					log.Printf("mysql: ensure schema failed: %v", err)
					_ = candidate.Close()
					candidate = nil
				} else {
					schemaCancel()
					log.Printf("mysql: connected and schema ready")
				}
			}
		}
		if candidate == nil && cfg.DBRequired {
			log.Fatalf("mysql: DB_REQUIRED=true but database unavailable")
		}
		store = candidate
	}
	if store != nil {
		defer store.Close()
	}

	// 3. 初始化 Redis（降级策略：连接失败不致命）
	cache, err := storage.NewRedisClient(cfg.Redis, cfg.Cache)
	if err != nil {
		log.Printf("redis: init failed (degraded mode): %v", err)
	}
	if cache != nil {
		defer cache.Close()
	}
	if cache == nil || !cache.IsEnabled() {
		log.Printf("redis: disabled or unavailable, serving without cache")
	} else {
		log.Printf("redis: connected")
	}

	// 4. 组装并启动 HTTP 服务
	apiServer := api.NewServer(store, cache, cfg.CORSOrigins)

	httpServer := &http.Server{
		Addr:         cfg.Addr,
		Handler:      apiServer.Routes(),
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	}

	// 优雅启停
	go func() {
		log.Printf("listening on http://127.0.0.1%s", cfg.Addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("http server failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutting down...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("http server shutdown error: %v", err)
	}
	log.Println("server stopped")
}
