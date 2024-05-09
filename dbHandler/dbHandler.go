package dbHandler

import (
	"bytes"
	"database/sql"
	"fmt"
	"log"
	"os"
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

	reader1, writer1, _ := os.Pipe()
	reader2, writer2, _ := os.Pipe()
	outBuff := bytes.Buffer{}
	//Each writer writes to its reader, they're connected through the pipe.

	cmd1 := exec.Command("/bin/bash", "-c", "psql -lqt")
	// cmd1 := exec.Command("/bin/bash", "-c", "echo hi")
	cmd1.Stdout = writer1
	err = cmd1.Start()
	if err != nil {
		log.Fatalf("Issue with starting the command %s: %v\n",
			cmd1.String(), err)
	}
	err = cmd1.Wait()
	writer1.Close()
	if err != nil {
		log.Fatalf("Issue with finishing the command %s: %v\n",
			cmd1.String(), err)
	}

	cmd2 := exec.Command("/bin/bash", "-c", `cut -d \| -f 1`)
	cmd2.Stdin = reader1
	cmd2.Stdout = writer2
	err = cmd2.Run()
	writer2.Close()
	if err != nil {
		log.Fatalf("Issue with running the command %s: %v\n",
			cmd2.String(), err)
	}

	cmd3String := fmt.Sprintf(`grep -qw '%s'`, dbName)
	cmd3 := exec.Command("/bin/bash", "-c", cmd3String)
	cmd3.Stdin = reader2
	cmd3.Stdout = &outBuff
	err = cmd3.Run()
	if err != nil {
		log.Fatalf("Issue with running the command %s: %v\n",
			cmd3.String(), err)
	}

	exitVal := outBuff.String()
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
