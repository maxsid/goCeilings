package main

import (
	"github.com/maxsid/goCeilings/api"
	"github.com/maxsid/goCeilings/api/storage"
	"log"
	"math/rand"
	"time"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	config := parseFlags()
	if config.PasswordSalt != "" {
		storage.HashSalt = config.PasswordSalt
	}
	if config.JWTSecret != "" {
		api.SigningSecret = config.JWTSecret
	}

	st, err := storage.NewSQLiteStorage(config.SQLiteFile)
	if err != nil {
		log.Fatalln(err)
	}
	if err := st.CreateAdmin(config.ForceAdmin); err != nil {
		log.Fatalln(err)
	}
	if err := api.Run(config.APIAddress, st); err != nil {
		log.Fatalln(err)
	}
}
