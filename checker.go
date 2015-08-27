package main

import (
	"os"
	"fmt"
	"encoding/csv"
)

func processFile(path string) {
	// open input file
	csvFile, err := os.Open(path);

	if err != nil {
		fmt.Printf("Cannot open file %s because of %q\n", path, err)

		os.Exit(1)
	}

	csvReader := csv.NewReader(csvFile)
	csvReader.Comma = ';'

	lines, err := csvReader.ReadAll()

	if err != nil {
		fmt.Printf("Error during reading of csv file %q\n", err)

		os.Exit(1)
	}

	// close file
	csvFile.Close()

	for index, line := range lines {
		fmt.Printf("Line %d %s %s\n", index, line[0], line[1])
	}
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

	processFile(inputFile)
}