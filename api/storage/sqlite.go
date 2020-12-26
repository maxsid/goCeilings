package storage

import "gorm.io/driver/sqlite"

func NewSQLiteStorage(name string) (*Storage, error) {
	return newDatabase(sqlite.Open(name))
}
