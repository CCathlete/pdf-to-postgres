package dbHandler

import (
	"bytes"
	"database/sql"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	pdfhandler "pdf-to-postgres/pdfHandler"
	"strings"
	"sync"

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
	dbName = strings.ToLower(dbName)

	reader1, writer1, _ := os.Pipe()
	reader2, writer2, _ := os.Pipe()
	var outBuff bytes.Buffer
	var wg sync.WaitGroup

	wg.Add(2)
	//Each writer writes to its reader, they're connected through the pipe.

	cmd1 := exec.Command("/bin/psql", "-lqt")
	// cmd1 := exec.Command("/bin/echo", "-e", `foo|bar\nbaz|quux\nfrob|fiddle\n`)
	cmd1.Stdout = writer1

	cmd2 := exec.Command("/bin/cut", "-d", "|", "-f", "1")
	cmd2.Stdin = reader1
	cmd2.Stdout = writer2

	cmd3 := exec.Command("/bin/grep", "-w", dbName)
	// cmd3 := exec.Command("/bin/grep", "-w", "Parasites")
	// cmd3 := exec.Command("/bin/grep", "-w", "baz")
	cmd3.Stdin = reader2
	cmd3.Stdout = &outBuff //If the db exists I'm expecting its name.

	go func() {
		defer wg.Done()
		err = cmd1.Run()
		if err != nil {
			log.Fatalf("Issue with running the command %s: %v\n",
				cmd1.String(), err)
		}

		err = writer1.Close()
		if err != nil {
			log.Fatalf("Issue with closing writer 1: %v\n", err)
		}
	}()

	go func() {
		defer wg.Done()
		err = cmd2.Run()
		if err != nil {
			log.Fatalf("Issue with running the command %s: %v\n",
				cmd2.String(), err)
		}

		err = writer2.Close()
		if err != nil {
			log.Fatalf("Issue with closing writer 2: %v\n", err)
		}

		_, err = io.Copy(os.Stdout, reader1)
		if err != nil {
			log.Fatalf("Issue with reading reader 1: %v\n", err)
		}
		err = reader1.Close()
		if err != nil {
			log.Fatalf("Issue with closing reader 1: %v\n", err)
		}
	}()

	err = cmd3.Run()
	if err != nil {
		log.Printf("Issue with running the command %s: %v. Database probably does not exist.\n",
			cmd3.String(), err)
	}
	grepOutput := outBuff.String()

	wg.Wait()

	err = reader2.Close()
	if err != nil {
		log.Fatalf("Issue with closing reader 2: %v\n", err)
	}

	fmt.Printf("grepOutput: %v", grepOutput)
	if len(grepOutput) == 0 {

		//If the database doesn't exist, we create it.
		query := fmt.Sprintf(`CREATE DATABASE %s;`, dbName)

		_, err = dbPointer.Exec(query)
		if err != nil {
			log.Fatalf("Couldn't run the query %s: %v\n", query, err)
		}
	}

	//After making sure we have an existing DB, we create a new connection.
	connectionString = fmt.Sprintf("host=%s port=%d"+
		" user=%s password=%s dbname=%s sslmode=disable",
		dbInfo["host"].(string), dbInfo["port"].(int),
		dbInfo["user"].(string),
		dbInfo["password"].(string),
		dbName,
	)
	dbPointer, err = sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatalf("Couldn't connect to posgres server: %v\n", err)
	}

	if len(grepOutput) == 0 {
		CreateTable(dbPointer, dbName)
	}

	return dbPointer
}

func CreateTable(dbPointer *sql.DB, tableName string) {
	// We want to take the keys of the inner structure and make
	// them keys in our table.
	//We assume that we're already connected to the DB.
	innerStructure := pdfhandler.ParasiteInfo{}
	innerStructure.Init()
	query := fmt.Sprintf(
		"CREATE TABLE %s(\nIndex SERIAL PRIMARY KEY NOT NULL",
		tableName)
	for key, _ := range innerStructure {
		query += fmt.Sprintf(",\n%s TEXT", strings.Replace(key,
			" ", "_", -1))
	}
	query += ");"
	_, err := dbPointer.Exec(query)
	if err != nil {
		log.Fatalf("Couldn't run the query %s: %v\n", query, err)
	}
}

func AddToTable(dbPointer *sql.DB, tableName string,
	entry pdfhandler.ParasiteInfo) {
	columns := []string{}
	values := []interface{}{} // Best practice for SQL in Go.
	placeholders := []string{}

	for key, value := range entry {
		// Sanitation of the keys before using SQL.
		key = strings.Replace(key, " ", "_", -1)
		key = strings.ToLower(key)
		// Collecting the column names and the corresponding values.
		columns = append(columns, key)
		values = append(values, value)
	}

	// Collecting placeholders in a slice.
	for i := 0; i < len(values); i++ {
		placeholders = append(placeholders,
			fmt.Sprintf("$%d", i+1))
	}
	// Constructing the query using the keys (columns) and placeholders.
	// Note: this is where the slices are each joined into one string with ", ".
	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s);",
		tableName,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "),
	)
	_, err := dbPointer.Exec(query, values...)
	if err != nil {
		log.Fatalf("Couldn't run the query %s: %v\n", query, err)
	}
}
