package main

import (
	"github.com/lionslon/go-keepass/internal/crypt"
	"github.com/lionslon/go-keepass/internal/logger"
	"github.com/lionslon/go-keepass/internal/server/app"
	"github.com/lionslon/go-keepass/internal/server/config"
	"github.com/lionslon/go-keepass/internal/storage"
	"log"
)

func main() {
	// Инициализируем синглтон логера
	if err := logger.Initialize("info"); err != nil {
		log.Fatalf("cannot initialize logger: %s\n", err)
	}

	//Разбираем конфиг
	cfg, err := config.Create()
	if err != nil {
		log.Fatalf("cannot load config: %s\n", err)
	}

	// Инициализируем расшифровыватель аутентификационных данных пользователя на закрытом ключе сервера
	err = crypt.NewDecryptor(cfg.CryptoKey)
	if err != nil {
		log.Fatalf("cannot initialize server credentions decryptor: %s", err)
	}

	// База данных
	storage, err := storage.NewKeeperStorage(cfg.DataBaseDSN)
	if err != nil {
		log.Fatalf("cannot create db store: %s\n", err)
	}
	defer storage.Close()

	//Сервер
	app, err := app.Create(cfg, storage)
	if err != nil {
		logger.Error("cannot create app: %s", err)
		return
	}

	//Запуск
	go app.Run()

	logger.Info("Running server: address %s", cfg.Endpoint)

	<-app.ServerDone()

	if err := app.Shutdown(); err != nil {
		logger.Error("Server shutdown failed: %s", err)
	}

	logger.Info("Server has been shutdown")
}
