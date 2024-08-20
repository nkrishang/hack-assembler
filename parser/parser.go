package parser

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type InstructionType int

const (
	A_INSTRUCTION = iota
	C_INSTRUCTION
	L_INSTRUCTION
)

// Returns the type of the current instruction.
func instructionType(instruction string) InstructionType {
	
	if strings.HasPrefix(instruction, "@") {
		return A_INSTRUCTION	
	}
	
	if strings.HasPrefix(instruction, "(") {
		return L_INSTRUCTION	
	}

	return C_INSTRUCTION
}

// Returns the symbol of a @xxx address/A_INSTRUCTION or (xxx) label/L_INSTRUCTION
func symbol(instruction string) string {
	var res string

	if strings.HasPrefix(instruction, "@") {
		res, _ = strings.CutPrefix(instruction, "@")
	}

	
	if strings.HasPrefix(instruction, "(") {
		temp, _ := strings.CutPrefix(instruction, "(")
		res, _ = strings.CutSuffix(temp, ")")
	}

	return res
}

// Returns storage destination a computation / C_INSTRUCTION
func dest(instruction string) string {
	if !strings.Contains(instruction, "=") {
		return "null"
	}

	dst, _, _ := strings.Cut(instruction, "=")
	return dst
}

// Returns the computation to perform of a C_INSTRUCTION
func comp(instruction string) string {
	beforeEq, afterEq, sep := strings.Cut(instruction, "=")

	cmp := ""

	if sep {
		cmp, _, _ = strings.Cut(afterEq, ";")
	} else {
		cmp, _, _ = strings.Cut(beforeEq, ";")
	}
	
	return cmp
}

// Returns where to jump based on a computation / C_INSTRUCTION
func jump(instruction string) string {
	if !strings.Contains(instruction, ";") {
		return "null"
	}

	_, jmp, _ := strings.Cut(instruction, ";")
	return  jmp
}

func asmToBinary(instruction string) string {

	switch instructionType(instruction) {

	case A_INSTRUCTION:
		addr := symbol(instruction)
		return DecimalStrToBinary15BitStr(addr)		

	case C_INSTRUCTION: 
		var result strings.Builder

		result.WriteString("111")
		result.WriteString(GetCompBinary(comp(instruction)))
		result.WriteString(GetDestBinary(dest(instruction)))
		result.WriteString(GetJumpBinary(jump(instruction)))

		return result.String()

	default:
		panic("Unsupported instruction type")
	}
}

func initializeSymbolTable() map[string]string {
	symbolTable := make(map[string]string)
	
	for i := range 16 {
		sym := strconv.FormatUint(uint64(i), 10)
		symbolTable["R" + sym] = sym
	}
	symbolTable["SP"] = "0"
	symbolTable["LCL"] = "1"
	symbolTable["ARG"] = "2"
	symbolTable["THIS"] = "3"
	symbolTable["THAT"] = "4"
	symbolTable["SCREEN"] = "16384"
	symbolTable["KBD"] = "24576"

	return symbolTable
} 

func Parse(inputFile, outputFile *os.File) {

	// Initialize symbolTable: R0=0, R1=1, ... etc.
	symbolTable := initializeSymbolTable()
	
	// First pass: identify label symbols and assign them RAM/ROM locations
	programCounter := 0

	fileScannerFirstPass := bufio.NewScanner(inputFile)
	for fileScannerFirstPass.Scan() {

		// Get and sanitize ASM line
		line := strings.TrimSpace(fileScannerFirstPass.Text())
		if line == "" || strings.HasPrefix(line, "/") {
			continue
		}
		line = strings.Split(line, " ")[0]

		if instructionType(line) == L_INSTRUCTION {
			symbolTable[symbol(line)] = strconv.Itoa(programCounter)
		} else {
			programCounter++
		}
	}

	// Check for errors during scanning
	if err := fileScannerFirstPass.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input file: %v\n", err)
		os.Exit(1)
	}

	// Reset the file pointer to the beginning of the file
	_, err := inputFile.Seek(0, 0)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error seeking input file: %v\n", err)
		os.Exit(1)
	}

	// Second pass: translate symbolized ASM to Binary
	ramCounter := 16
	
	fileScannerSecondPass := bufio.NewScanner(inputFile)
	for fileScannerSecondPass.Scan() {

		// Get and sanitize ASM line
		line := strings.TrimSpace(fileScannerSecondPass.Text())
		if line == "" || strings.HasPrefix(line, "/") {
			continue
		}
		line = strings.Split(line, " ")[0]
		
		s := symbol(line)
		var nonSymbolizedLine string

		switch instructionType(line) {
			case C_INSTRUCTION:				
				nonSymbolizedLine = line

			case A_INSTRUCTION:
				if _, err := strconv.Atoi(s); err != nil {
					if symbolTable[s] == "" {
						symbolTable[s] = strconv.Itoa(ramCounter)	
						ramCounter++
					}
					nonSymbolizedLine = strings.Replace(line, s, symbolTable[s], 1)
				} else {
					nonSymbolizedLine = line
				}

			default:
				continue
		}

		// Translate to binary
		parsed := asmToBinary(nonSymbolizedLine)

		// Write binary to output file
		_, err := outputFile.WriteString(parsed + "\n")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing to output file: %v\n", err)
			os.Exit(1)
		}

	}

	// Check for errors during scanning
	if err := fileScannerSecondPass.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input file: %v\n", err)
		os.Exit(1)
	}
}