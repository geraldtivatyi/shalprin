package profile

import (
	"database/sql"
	"fmt"
	"regexp"

	"github.com/oligoden/chassis/device/model/data"
)

type Record struct {
	Firstname string `form:"firstname" json:"firstname"`
	Lastname  string `form:"lastname" json:"lastname"`
	Email     string `form:"email" json:"email"`
	data.Default
}

func NewRecord() *Record {
	r := &Record{}
	r.Default = data.Default{}
	r.Perms = ":::c"
	return r
}

func (e Record) Prepare() error {
	match, _ := regexp.MatchString("^[a-zA-Z][a-zA-Z- ]{0,20}[a-zA-Z]$", e.Firstname)
	if !match {
		return fmt.Errorf("bad request, firstname not valid")
	}

	match, _ = regexp.MatchString("^[a-zA-Z][a-zA-Z- ]{0,20}[a-zA-Z]$", e.Lastname)
	if !match {
		return fmt.Errorf("bad request, lastname not valid")
	}

	match, _ = regexp.MatchString("(?:[a-z0-9!#$%&'*+/=?^_`{|}~-]+(?:\\.[a-z0-9!#$%&'*+/=?^_`{|}~-]+)*|\"(?:[\x01-\x08\x0b\x0c\x0e-\x1f\x21\x23-\x5b\x5d-\x7f]|\\[\x01-\x09\x0b\x0c\x0e-\x7f])*\")@(?:(?:[a-z0-9](?:[a-z0-9-]*[a-z0-9])?\\.)+[a-z0-9](?:[a-z0-9-]*[a-z0-9])?|\\[(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?|[a-z0-9-]*[a-z0-9]:(?:[\x01-\x08\x0b\x0c\x0e-\x1f\x21-\x5a\x53-\x7f]|\\[\x01-\x09\x0b\x0c\x0e-\x7f])+)\\])", e.Email)
	if !match {
		return fmt.Errorf("bad request, email not valid")
	}

	// match, _ = regexp.MatchString("^\\+?[0-9 ]{10,14}$", e.Mobile)
	// if !match {
	// 	return fmt.Errorf("bad request, mobile not valid")
	// }

	// match, _ = regexp.MatchString("^[a-zA-Z0-9- ]{10,}$", e.Address)
	// if !match {
	// 	return fmt.Errorf("bad request, address not valid")
	// }

	// match, _ = regexp.MatchString("^(?=.*[A-Za-z])(?=.*\\d)(?=.*[@$!%*#?&])[A-Za-z\\d@$!%*#?&]{8,}$", e.Password)
	// if !match {
	// 	return fmt.Errorf("bad request, password not valid")
	// }

	return nil
}

func (Record) TableName() string {
	return "profiles"
}

func (Record) Migrate(db *sql.DB) error {
	q := "CREATE TABLE `profiles` ( `firstname` varchar(255) DEFAULT NULL, `lastname` varchar(255) DEFAULT NULL, `email` varchar(255) DEFAULT NULL, `id` int(10) unsigned NOT NULL AUTO_INCREMENT, `uc` varchar(255) DEFAULT NULL, `owner_id` int(10) unsigned DEFAULT NULL, `perms` varchar(255) DEFAULT NULL, `hash` varchar(255) DEFAULT NULL, PRIMARY KEY (`id`), UNIQUE KEY `uc` (`uc`) )"

	_, err := db.Exec(q)
	if err != nil {
		return fmt.Errorf("doing migration: %w", err)
	}
	return nil
}
