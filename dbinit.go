package main

import (
	_ "embed"
	"fmt"
)

//go:embed rblrdb.sql
var initsql string

func checkDB() bool {

	sqlx := "SELECT DBInitialised FROM config"
	rows, err := DBH.Query(sqlx)
	if err != nil {
		return false
	}
	defer rows.Close()
	var dbi string
	if !rows.Next() {
		return false
	}
	err = rows.Scan(&dbi)
	if err != nil {
		return false
	}
	return dbi == "1"

}

func createDB() {

	fmt.Println("Initialising database")
	_, err := DBH.Exec(initsql)
	checkerr(err)

}
