package data

import (
	"database/sql"
	"log"
)

var db *sql.DB

func NewDb(base *sql.DB) {
	db = base
}

func CreateTable() {
	query := `create table if not exists list(
		id serial primary key,
		ch varchar(20) 
	)`
	_, err := db.Exec(query)
	if err != nil {
		log.Fatal(err)
	}

}

func GetNum() (int64, error) {
	num, err := CheckDB()
	if err != nil {
		return -1, err
	}
	query := "insert into list (ch) values($1)"
	_, err = db.Exec(query, "1")
	if err != nil {
		return -1, err
	}
	return num, nil
}

func CheckDB() (int64, error) {
	var count int64
	query := `select count(id) from list`
	res, err := db.Query(query)
	if err != nil {
		return -1, err
	}
	defer res.Close()
	for res.Next() {
		err := res.Scan(&count)
		if err != nil {
			return -1, err
		}
	}
	return count, nil
}
