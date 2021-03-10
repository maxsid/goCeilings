package sqlite

import (
	"github.com/maxsid/goCeilings/server/common/storage/gorm"
	"gorm.io/driver/sqlite"
)

func NewSQLiteStorage(name string) (*gorm.Storage, error) {
	return gorm.NewDatabase(sqlite.Open(name))
}
