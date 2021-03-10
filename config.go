package main

import "flag"

const (
	defaultAPIAddress = "127.0.0.1:8081"
	defaultSQLiteFile = "go-ceiling.db"
)

type Config struct {
	APIAddress   string
	SQLiteFile   string
	JWTSecret    string
	PasswordSalt string
	ForceAdmin   bool
}

func parseFlags() *Config {
	config := Config{}
	flag.StringVar(&config.APIAddress, "addr", defaultAPIAddress, "address of API for listening")
	flag.StringVar(&config.SQLiteFile, "sqlite", defaultSQLiteFile, "file of SQLite data storage")
	flag.StringVar(&config.JWTSecret, "secret", "", "secret for singing of jwt token  (default value in the code)")
	flag.StringVar(&config.PasswordSalt, "salt", "", "salt for users passwords  (default value in the code)")
	flag.BoolVar(&config.ForceAdmin, "admin", false, "create a new administrator "+
		"(the app create an admin user automatically if database doesn't have at least one)")
	flag.Parse()
	return &config
}
