package storage

import (
	"database/sql"
	"fmt"
	"github.com/pressly/goose"
	"log"
	"os"
)

func Migrate(config *MysqlConfig) {
	db, err := sql.Open("mysql", config.dsn())
	if err != nil {
		panic(err)
	}
	defer db.Close()
	if err := goose.SetDialect("mysql"); err != nil {
		panic(err)
	}
	migrationDir := os.Getenv("MIGRATION_DIR")
	log.Printf("migration dir is: %v\n", migrationDir)
	if err := goose.Status(db, migrationDir); err != nil {
		panic(err)
	}
	if err := goose.Up(db, migrationDir); err != nil {
		panic(err)
	}
}

func CreateDatabase(config *MysqlConfig, drop bool) {
	db, err := sql.Open("mysql",
		fmt.Sprintf("%s:%s@tcp(%s:3306)/",
			config.username, config.password, config.host),
	)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			fmt.Printf("error closing db: %v", err)
		}
	}()
	if drop {
		if _, err = db.Exec("drop database if exists " + config.dbName); err != nil {
			panic(err)
		}
	}
	if _, err = db.Exec("create database if not exists " + config.dbName); err != nil {
		panic(err)
	}
}
