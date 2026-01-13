package db

import (
	"database/sql"
	"fmt"
	"strconv"

	util "telegramconnect/util"

	_ "github.com/lib/pq"
)

var db *sql.DB

func Connect() (string, int, error) {
	dbVars, err := util.ValueGetter("DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD", "DB_NAME")
	if err != nil {
		return "Failed to get database variables", 500, err
	}
	host := dbVars["DB_HOST"]
	portStr := dbVars["DB_PORT"]
	user := dbVars["DB_USER"]
	pass := dbVars["DB_PASSWORD"]
	dbname := dbVars["DB_NAME"]

	// Convert port to int if needed
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return "Failed to convert port to int", 500, err
	}

	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, pass, dbname)

	var dbErr error
	db, dbErr = sql.Open("postgres", connStr)
	if dbErr != nil {
		return "Failed to open database connection", 500, dbErr
	}

	err = db.Ping()
	if err != nil {
		return "Failed to ping database", 500, err
	}

	return "Connected to database", 200, nil
}

func Disconnect() {
	if db != nil {
		db.Close()
		db = nil
	}
}

func Query(query string) {
	// Code to execute a database query
}

func Insert(data interface{}) {
	// Code to insert data into the database
}

func Update(data interface{}) {
	// Code to update data in the database
}

func Delete(id string) {
	// Code to delete data from the database
}
