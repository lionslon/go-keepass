// Package config предназначен для инициализации конфигурации клиента.
package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"
)

// Config содержит список параметров для работы клиента.
type Config struct {
	ServerEndpoint string //эндпонт сервера
	CryptoKey      string //путь до файла с публичным ключом сервера для шифрования логина и пароля (карманный tls)
	ConfigJson     string //путь до файла с json конфигурацией
	PollInterval   int64  //интервал обновления данных
}

// formJson дополняет отсутствующие параметры из json
func (m *Config) formFile() error {

	data, err := os.ReadFile(m.ConfigJson)
	if err != nil {
		return fmt.Errorf("cannot read json config: %w", err)
	}

	var settings map[string]interface{}

	err = json.Unmarshal(data, &settings)
	if err != nil {
		return fmt.Errorf("cannot unmarshal json settings: %w", err)
	}

	for stype, value := range settings {
		switch stype {
		case "address":
			if m.ServerEndpoint == `` {
				m.ServerEndpoint = value.(string)
			}
		case "poll_interval":
			if m.PollInterval == 0 {
				duration, err := time.ParseDuration(value.(string))
				if err != nil {
					return fmt.Errorf("bad json param 'poll_interval': %w", err)
				}
				m.PollInterval = int64(duration.Seconds())
			}
		case "crypto_key":
			if m.CryptoKey == `` {
				m.CryptoKey = value.(string)
			}
		}
	}

	return nil
}

// LoadAgentConfig загружает настройки клиента из командной строки или переменных окружения.
func LoadAgentConfig() (*Config, error) {
	cfg := new(Config)
	/*Получаем параметры из командной строки*/
	flag.StringVar(&cfg.ServerEndpoint, "a", "localhost:8033", "server address and port")
	flag.Int64Var(&cfg.PollInterval, "p", 10, "poll interval")
	flag.StringVar(&cfg.CryptoKey, "k", "public.rsa", "open crypt key")
	flag.StringVar(&cfg.ConfigJson, "c", "", "json config")

	flag.Parse()

	if cfg.ConfigJson != `` {
		err := cfg.formFile()
		if err != nil {
			return nil, fmt.Errorf("cannot full setting from json config: %w", err)
		}
	}

	return cfg, nil
}
