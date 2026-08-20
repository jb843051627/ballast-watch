package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ballast-watch/internal/collector"
	"ballast-watch/internal/config"
	"ballast-watch/internal/handler"
	"ballast-watch/internal/service"
	"ballast-watch/internal/store"
)

func main() {
	cfg := config.Load()
	if err := cfg.SmokeCheck(); err != nil {
		log.Fatalf("配置校验失败: %v", err)
	}
	if err := config.EnsureDBDir(cfg.DBPath); err != nil {
		log.Fatalf("准备数据库目录失败: %v", err)
	}

	db, err := store.Open(cfg.DBPath)
	if err != nil {
		log.Fatalf("打开数据库失败: %v", err)
	}
	defer db.Close()

	svc := service.NewServices(db, cfg)

	col := collector.NewCollector(cfg, svc)
	col.Start()
	defer col.Stop()

	if cfg.EnableSimulator {
		sim := collector.NewSimulator(cfg, store.NewSamplingPointStore(db), store.NewSensorStore(db), col)
		sim.Start()
	}

	h := handler.NewHandler(svc, cfg.WebDir)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go backgroundTasks(ctx, svc, cfg)

	srv := &http.Server{
		Addr:              ":" + cfg.ServerPort,
		Handler:           h.Router(),
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("ballast-watch 启动: http://localhost:%s", cfg.ServerPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP 服务异常: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	log.Println("收到退出信号，优雅停机...")
	sctx, scancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer scancel()
	_ = srv.Shutdown(sctx)
}

// backgroundTasks 后台周期任务：看板快照刷新 + 传感器离线标记。
func backgroundTasks(ctx context.Context, svc *service.Services, cfg *config.AppConfig) {
	snapshotTicker := time.NewTicker(cfg.SnapshotRefresh)
	defer snapshotTicker.Stop()
	offlineTicker := time.NewTicker(30 * time.Second)
	defer offlineTicker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-snapshotTicker.C:
			if err := svc.Dashboard.Refresh(ctx); err != nil {
				log.Printf("看板刷新失败: %v", err)
			}
		case <-offlineTicker.C:
			if err := svc.Sensors.MarkOffline(ctx, time.Now(), cfg.OfflineAfter); err != nil {
				log.Printf("传感器离线检查失败: %v", err)
			}
		}
	}
}