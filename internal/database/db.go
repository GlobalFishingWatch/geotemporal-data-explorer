package database

import (
	"database/sql"

	"github.com/4wings/cli/types"
	"github.com/spf13/viper"
)

type db struct {
	db *sql.DB
}

var LocalDB types.LocalDatabase

// var BQDB types.Database

func Open() (err error) {

	if viper.GetBool("local") {
		LocalDB, err = openDuckDB()
		if err != nil {
			return err
		}
	}

	// BQDB, err = openBQ()
	// if err != nil {
	// 	return err
	// }
	return nil
}

func Close() {
	if viper.GetBool("local") {
		LocalDB.Close()
	}
	// BQDB.Close()
}
