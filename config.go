package main

import (
	"flag"
	"os"
	"strconv"
)

type Config struct {
	Port int
}

func LoadConfig() Config {
	cfg := Config{}
	flag.IntVar(&cfg.Port, "port", 8085, "Port to listen on")
	flag.Parse()
	return cfg
}

func LoadConfigFromEnv() Config {
	cfg := LoadConfig()

	port, ok := os.LookupEnv("PORT")
	if ok {
		if p, err := strconv.Atoi(port); err == nil {
			cfg.Port = p
		}
	}

	return cfg
}
