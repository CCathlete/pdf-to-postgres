package dbhandler

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/jackc/pgx"
)

func DbInit(dbInfo map[string]interface{}) *sql.DB {
	connectionString := fmt.Sprintf("host=%s port=%d"+
		"user=%s password=%s sslmode=disable",
		dbInfo["host"].(string), dbInfo["port"].(int),
		dbInfo["user"].(string),
		dbInfo["password"].(string),
	)
	dbPointer, err := sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatalf("Couldn't connect to posgres server: %v\n", err)
	}
	dbName := dbInfo["name"].(string)

	//If the database doesn't exist, we create it.
	query := fmt.Sprintf(
		`SELECT 'CREATE DATABASE %s' WHERE NOT EXISTS `+
			`(SELECT FROM pg_database WHERE datname = '%s')\gexec`,
		dbName, dbName)
	_, err = dbPointer.Exec(query)
	if err != nil {
		log.Fatalf("Couldn't run the query %s: %v\n", query, err)
	}

	dbPointer.Close()
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
