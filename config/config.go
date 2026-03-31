package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	DBConn        string
	RabbitURL     string
	NodeID        string
	FetchInterval time.Duration
	LockInterval  time.Duration
}

func Load() Config {
	fiStr := os.Getenv("FETCH_INTERVAL")
	fi, err := strconv.Atoi(fiStr)
	if err != nil {
		panic(err)
	}
	liStr := os.Getenv("LOCK_INTERVAL")
	li, err := strconv.Atoi(liStr)
	if err != nil {
		panic(err)
	}
	return Config{
		DBConn:        os.Getenv("DB_CONN"),
		RabbitURL:     os.Getenv("RABBIT_URL"),
		NodeID:        os.Getenv("NODE_ID"),
		FetchInterval: time.Duration(fi),
		LockInterval:  time.Duration(li),
	}
}
