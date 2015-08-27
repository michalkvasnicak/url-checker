package main

import "os"
import "fmt"


func main() {
	if len(os.Args) < 2 {
		fmt.Printf("Missing input file argument\n")

		os.Exit(1)
	}

	var inputFile string = os.Args[1];

	fmt.Printf("URL-checker\n");

	if _, err := os.Stat(inputFile); os.IsNotExist(err) {
		fmt.Printf("File %s does not exist", inputFile)

		os.Exit(1)
	}
}