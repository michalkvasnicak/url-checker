package main

import (
	"database/sql"
	"os"
	"fmt"
	"time"
	_ "github.com/mattn/go-sqlite3"
	"log"
)

type matchResult struct{
	url string
	matches []string
	pattern string
	created_at time.Time
}

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("Missing input file argument\n")
		log.Fatal()
	}

	var inputFile string = os.Args[1];

	fmt.Printf("URL-checker\n");

	if _, err := os.Stat(inputFile); os.IsNotExist(err) {
		fmt.Printf("File %s does not exist\n", inputFile)
		log.Fatal(err)
	}

	// remove test db
	os.Remove("./test.db")

	db, err := sql.Open("sqlite3", "./test.db")

	if err != nil {
		fmt.Printf("Cannot open sqlite3 database %q\n", err)
		log.Fatal(err)
	}

	// create database
	sqlStmt := `
		create table results (id integer not null primary key autoincrement, url TEXT, regexp TEXT, matches TEXT, created_at INTEGER);
		delete from results;
	`

	if _, err := db.Exec(sqlStmt); err != nil {
		log.Fatal(err)
	}

	manager := func(ticker <-chan time.Time) (chan time.Time) {
		out := make(chan time.Time)

		// start immediately
		go func() {
			out <- time.Now()
		}()

		// listen to ticker
		go func() {
			for t := range ticker {
				out <- t
			}
		}()

		return out
	}

	processChan := manager(time.NewTicker(time.Minute * 60).C)
	resultChan := make(chan *matchResult)

	for {
		select {
		case result := <- resultChan:
			fmt.Printf("Received result for %s with %d matches\n", result.url, len(result.matches))

			saveResult(result, db)
		case <- processChan:
			startedAt := time.Now()
			fmt.Printf("Processing started\n")

			lines, err := processFile(inputFile)

			if err != nil {
				fmt.Printf("Error during processing file %q %q\n", err, lines)

				log.Fatal(err)
			}

			// for every line start go coroutine
			for _, line := range lines {
				go processLine(line, startedAt, resultChan)
			}
		}
	}
}