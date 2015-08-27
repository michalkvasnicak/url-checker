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
	"strings"
)

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
