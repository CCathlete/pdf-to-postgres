package dbHandler

import (
	"bytes"
	"database/sql"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
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
	r1Info, _ := reader1.Stat()
	content := make([]byte, r1Info.Size())
	reader1.Read(content)
	fmt.Printf("%s", content)

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
	r2Info, _ := reader2.Stat()
	content = make([]byte, r2Info.Size())
	reader2.Read(content)
	fmt.Printf("%s", content)

	err = cmd3.Run()
	if err != nil {
		log.Printf("Issue with running the command %s: %v\n",
			cmd3.String(), err)
	}
	exitVal := outBuff.String()

	wg.Wait()

	err = reader2.Close()
	if err != nil {
		log.Fatalf("Issue with closing reader 2: %v\n", err)
	}

	fmt.Printf("exitVal: %v", exitVal)
	if len(exitVal) == 0 {

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
