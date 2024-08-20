package main

import (
	"bufio"
	"fmt"
	"hackassembler/parser"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	// Prompt user for file path
	fmt.Print("Enter the path to the file: ")

	// Read user input
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	filePath := scanner.Text()

	// Only accept ASM files
	if !strings.Contains(filePath, ".asm") {
		fmt.Fprintf(os.Stderr, "Error: cannot parse non-ASM files.\n")
		os.Exit(1)
	}

	// Open the input file
	inputFile, err := os.Open(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening input file: %v\n", err)
		os.Exit(1)
	}
	defer inputFile.Close()

	// Create the output directory if it doesn't exist
	outputDir := "output"
	if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output directory: %v\n", err)
		os.Exit(1)
	}

	// Create the .hack output file in the output directory
	outputFileName := strings.Split(filepath.Base(filePath), ".")[0] + ".hack"
	outputFilePath := filepath.Join(outputDir, outputFileName)
	outputFile, err := os.Create(outputFilePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output file: %v\n", err)
		os.Exit(1)
	}
	defer outputFile.Close()

	// Parse input file and write result in output file
	parser.Parse(inputFile, outputFile)

	fmt.Printf("Contents have been written to %s\n", outputFilePath)
}
