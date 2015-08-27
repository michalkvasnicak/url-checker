package main

import (
	"os"
	"fmt"
	"encoding/csv"
	"time"
)

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

	tickerChan := time.NewTicker(time.Minute * 60).C
	processChan := make(chan bool)

	go func() {
		processChan <- true // start processing
	}()

	for {
		select {
		case <- tickerChan:
			processChan <- true
		case <- processChan:
			fmt.Printf("Processing started\n")

			lines, err := processFile(inputFile)

			if err != nil {
				fmt.Printf("Error during processing file %q %q\n", err, lines)

				os.Exit(1)
			}
		}
	}
}