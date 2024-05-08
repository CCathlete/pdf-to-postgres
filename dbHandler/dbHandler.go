package dbHandler

import (
	"database/sql"
	"fmt"
	"log"
	"os/exec"

	_ "github.com/lib/pq"
)

func DbInit(dbInfo map[interface{}]interface{}) *sql.DB {
	//Creating a connection withouot a specific DB.
	connectionString := fmt.Sprintf("host=%s port=%d"+
		" user=%s password=%s sslmode=disable",
		dbInfo["host"].(string), dbInfo["port"].(int),
		dbInfo["user"].(string),
		dbInfo["password"].(string),
	)
	dbPointer, err := sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatalf("Couldn't connect to posgres server: %v\n", err)
	}

	dbName := dbInfo["name"].(string)

	queryString := fmt.Sprintf(
		`psql -lqt | cut -d \| -f 1 | grep -qw '%s'; echo $?)`,
		dbName)
	exitVal, _ := exec.Command(queryString).Output()

	if len(exitVal) != 0 {

		//If the database doesn't exist, we create it.
		query := fmt.Sprintf(`CREATE DATABASE %s;`, dbName)

		_, err = dbPointer.Exec(query)
		if err != nil {
			log.Fatalf("Couldn't run the query %s: %v\n", query, err)
		}
	}

	//After making sure we have an existing DB, we create a new connection.
	connectionString = fmt.Sprintf("host=%s port=%d"+
		"user=%s password=%s dbname=%s sslmode=disable",
		dbInfo["host"].(string), dbInfo["port"].(int),
		dbInfo["user"].(string),
		dbInfo["password"].(string),
		dbName,
	)
	dbPointer, err = sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatalf("Couldn't connect to posgres server: %v\n", err)
	}

	return dbPointer
}
