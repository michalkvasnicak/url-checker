package main

import (
	"os"
	"fmt"
	"encoding/csv"
	"time"
	"net/http"
	"regexp"
	"io/ioutil"
)

func fetchAndMatchAgainstUrlContent(url, pattern string) ([]string, error) {
	matcher, err := regexp.Compile(pattern)

	if err != nil {
		return nil, err
	}

	response, err := http.Get(url)

	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	content, err := ioutil.ReadAll(response.Body)

	if err != nil {
		return nil, err
	}

	return matcher.FindAllString(string(content), -1), nil
}

func processLine(line []string) {
	fmt.Printf("Processing %s against pattern %s\n", line[0], line[1])

	matches, err := fetchAndMatchAgainstUrlContent(line[0], line[1])

	fmt.Printf("Processed %s, matched %s\n", line[0], matches)

	if err != nil {
		fmt.Printf("Error while processing line %q\n", err)

		return
	}
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

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("Missing input file argument\n")

		os.Exit(1)
	}

	var inputFile string = os.Args[1];

	fmt.Printf("URL-checker\n");

	if _, err := os.Stat(inputFile); os.IsNotExist(err) {
		fmt.Printf("File %s does not exist\n", inputFile)

		os.Exit(1)
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

	for {
		select {
		case <- processChan:
			fmt.Printf("Processing started\n")

			lines, err := processFile(inputFile)

			if err != nil {
				fmt.Printf("Error during processing file %q %q\n", err, lines)

				os.Exit(1)
			}

			// for every line start go coroutine
			for _, line := range lines {
				go processLine(line)
			}
		}
	}
}