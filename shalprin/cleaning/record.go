package cleaning

import (
	"database/sql"
	"fmt"

	"github.com/oligoden/chassis/device/model/data"
	"github.com/oligoden/chassis/storage"
)

type Record struct {
	data.Default
}

func NewRecord() *Record {
	r := &Record{}
	r.Default = data.Default{}
	r.Perms = ":::c"
	return r
}

func (Record) TableName() string {
	return "cleaning"
}

func (e *Record) Read(db storage.DBReader, params ...string) error {
	db.First(e, params...)
	return nil
}

func (Record) Migrate(db *sql.DB) error {
	q := "CREATE TABLE `cleaning` ( `id` int(10) unsigned NOT NULL AUTO_INCREMENT, `uc` varchar(255) DEFAULT NULL, `owner_id` int(10) unsigned DEFAULT NULL, `perms` varchar(255) DEFAULT NULL, `hash` varchar(255) DEFAULT NULL, PRIMARY KEY (`id`), UNIQUE KEY `uc` (`uc`) )"

	_, err := db.Exec(q)
	if err != nil {
		return fmt.Errorf("doing migration: %w", err)
	}
	return nil
}
