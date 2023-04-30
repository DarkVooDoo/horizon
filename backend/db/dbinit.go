package db

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

const (
	SQL_C_STRING = "postgresql://momo:1234567@db:5432/momo?sslmode=disable"
)

//91mLLi9nTEGAGE4u

var db *sql.DB = nil

func GetDbInstance() *sql.DB {
	if db != nil {
		return db
	} else {
		dbInstance, err := sql.Open("postgres", SQL_C_STRING)
		if err != nil {
			fmt.Println(err)
		}
		db = dbInstance
		return dbInstance
	}
}
