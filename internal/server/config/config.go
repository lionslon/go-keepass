package config

import (
	"flag"
	"fmt"
	"os"

	"time"
)

type Config struct {
	Endpoint    string        `env:"RUN_ADDRESS"`
	DataBaseDSN string        `env:"DATABASE_DSN"`
	CryptoKey   string        `env:"RUN_ADDRESS"`  //Путь до файла с приватным ключом сервера для расшифровывания данных
	JWTKey      []byte        `env:"JWT_KEY"`      //Ключ для создания/проверки jwt для авторизации
	JWTDuration time.Duration `env:"JWT_DURATION"` //Время действия jwt для авторизации
}

func Create() (*Config, error) {
	cfg := &Config{}
	var JWTKey, JWTDuration string
	flag.StringVar(&cfg.Endpoint, "a", "localhost:8088", "address and port to run server")
	flag.StringVar(&cfg.DataBaseDSN, "d", "", "db dsn")
	flag.StringVar(&cfg.CryptoKey, "p", "private.rsa", "Server private key path")
	flag.StringVar(&JWTKey, "k", "gBz65sbl0GAb", "JWT key")
	flag.StringVar(&JWTDuration, "t", "60m", "JWT duration")
	flag.Parse()

	if cfg.DataBaseDSN == `` {
		return nil, fmt.Errorf("db dsn is empty")
	}

	if duration, exist := os.LookupEnv("JWT_DURATION"); exist {
		JWTDuration = duration
	}
	cfg.JWTKey = []byte(JWTKey)
	if duration, err := time.ParseDuration(JWTDuration); err != nil {
		return nil, fmt.Errorf("JWT DURATION: %w", err)
	} else {
		cfg.JWTDuration = duration
	}

	return cfg, nil
}
