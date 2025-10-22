package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/IvanChernomyrdin/avito-key-generate/config"
	"github.com/IvanChernomyrdin/avito-key-generate/config/db"
	"github.com/IvanChernomyrdin/avito-key-generate/internal/handler"
	"github.com/IvanChernomyrdin/avito-key-generate/runtime/logger"
)

func main() {
	//получили конфиг
	cfg := config.NewConfig()
	//создаём логирование
	logger, err := logger.NewLogger(cfg.LogDir, cfg.LogFileMaxSize, logger.DEBUG)
	if err != nil {
		panic(err)
	}
	defer logger.Close()
	logger.Info("main", "The logger is initialized")

	//подключаем базу данных
	database, err := db.NewConnectionPostgres(cfg.DatabaseDSN)
	if err != nil {
		logger.Fatal("database", "Connection failed", err)
	}
	defer db.CloseConnection(database)
	logger.Info("database", "Successfully connected to PostgreSQL")

	if err := db.MigrationsDatabase(database); err != nil {
		logger.Fatal("database", "Failed migrations", err)
	}
	logger.Info("database", "Migrations have been successfully applied")

	//подключаем роутер
	h := handler.NewHandler(database, logger)
	r := handler.NewRouter(h)

	//запускаем сервер
	logger.Info("server", "Starting HTTP server on "+cfg.Addr)

	server := &http.Server{
		Addr:    cfg.Addr,
		Handler: r,
	}
	//если сервер ляжет он сделает всё через graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		if err := server.ListenAndServe(); err != nil {
			logger.Error("server", "Failed to start server", err)
		}
	}()
	<-quit
	logger.Info("server", "Initiating a graceful shutdown of the server")

	// старт: выполнение каких-то функций перед завершением

	// конец: выполнение каких-то функций перед завершением

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Fatal("server", "Force server shutdown", err)
	}
	logger.Info("server", "graceful shudown completed")
}
