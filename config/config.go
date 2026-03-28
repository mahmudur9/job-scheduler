package config

import "os"

type Config struct {
	DBConn    string
	RabbitURL string
	NodeID    string
}

func Load() Config {
	return Config{
		DBConn:    os.Getenv("DB_CONN"),
		RabbitURL: os.Getenv("RABBIT_URL"),
		NodeID:    os.Getenv("NODE_ID"),
	}
}
