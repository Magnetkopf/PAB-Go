package main

import (
	"flag"

	"github.com/Magnetkopf/PAB-Go/internal/app"
	"github.com/Magnetkopf/PAB-Go/internal/config"
	jsonstore "github.com/Magnetkopf/PAB-Go/internal/store/json"
)

func main() {
	bind := flag.String("bind", "127.0.0.1:7212", "HTTP bind address (ip:port)")
	flag.Parse()

	cfg, err := config.LoadOrInit()
	if err != nil {
		panic(err)
	}
	store, err := jsonstore.New()
	if err != nil {
		panic(err)
	}
	if err := app.NewServer(cfg, store).Run(*bind); err != nil {
		panic(err)
	}
}
