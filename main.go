package main

import (
	"log"
	"math/rand"
	"time"

	"github.com/maxsid/goCeilings/server/api"
	"github.com/maxsid/goCeilings/server/common/storage/gorm"
	"github.com/maxsid/goCeilings/server/common/storage/sqlite"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	config := parseFlags()
	if config.PasswordSalt != "" {
		gorm.HashSalt = config.PasswordSalt
	}
	if config.JWTSecret != "" {
		api.SigningSecret = config.JWTSecret
	}

	st, err := sqlite.NewSQLiteStorage(config.SQLiteFile)
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
