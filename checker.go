package main

import (
	"database/sql"
	"os"
	"fmt"
	"encoding/csv"
	"time"
	"net/http"
	"regexp"
	"io/ioutil"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"strings"
)

type matchResult struct{
	url string
	matches []string
	pattern string
	created_at time.Time
}

func fetchAndMatchAgainstUrlContent(url, pattern string) ([]string, error) {
	matcher, err := regexp.Compile(pattern)

	if err != nil {
		return nil, err
	}

	response, err := http.Get(url)

	if err != nil {
		return nil,  err
	}

	defer response.Body.Close()

	content, err := ioutil.ReadAll(response.Body)

	if err != nil {
		return nil, err
	}

	return matcher.FindAllString(string(content), -1), nil
}

func processLine(line []string, startedAt time.Time, output chan *matchResult) {
	fmt.Printf("Processing %s against pattern %s\n", line[0], line[1])

	matches, err := fetchAndMatchAgainstUrlContent(line[0], line[1])

	fmt.Printf("Processed %s, matched %s\n", line[0], matches)

	if err != nil {
		fmt.Printf("Error while processing line %q\n", err)

		return
	}

	output <- &matchResult{line[0], matches, line[1], startedAt}
}

func processFile(path string) ([][]string, error) {
	// open input file
	csvFile, err := os.Open(path);

	if err != nil {
		fmt.Printf("Cannot open file %s because of %q\n", path, err)

		return nil, err
	}

	csvReader := csv.NewReader(csvFile)
	csvReader.Comma = ';'

	lines, err := csvReader.ReadAll()

	if err != nil {
		fmt.Printf("Error during reading of csv file %q\n", err)

		return nil, err
	}

	// close file
	csvFile.Close()

	return lines, nil
}

func saveResult(result *matchResult, db *sql.DB) {
	// find if there is same match
	stmt, err := db.Prepare("INSERT INTO results (url, regexp, matches, created_at) VALUES ($1, $2, $3, $4)")

	if err != nil {
		fmt.Printf("Error preparing query %q\n", err)
		return
	}

	if _, err := stmt.Query(result.url, result.pattern, strings.Join(result.matches, ""), result.created_at.Unix()); err != nil {
		fmt.Printf("Error inserting result to db %q\n", err)
		return
	}

	fmt.Printf("Result for %s and pattern %s saved\n", result.url, result.pattern)
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
		create table results (id integer not null primary key, url TEXT, regexp TEXT, matches TEXT, created_at INTEGER);
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