package parser

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// type Parser struct {
// 	fileScanner *bufio.Scanner
// 	hasMoreLines bool
// 	currentInstruction string
// }

type InstructionType int

const (
	A_INSTRUCTION = iota
	C_INSTRUCTION
	L_INSTRUCTION
)

// func InitParser(inputFile *os.File) *Parser {
// 	fileScanner := bufio.NewScanner(inputFile)
	
// 	hasMoreLines := fileScanner.Scan()
// 	currentInstruction := fileScanner.Text()

// 	return &Parser{
// 		fileScanner: fileScanner,
// 		hasMoreLines: hasMoreLines,
// 		currentInstruction: currentInstruction,
// 	}
// }

// // Advances to the next instruction. TODO: Handle an empty file
// func (p *Parser) advance() bool {

// 	instruction := p.currentInstruction

// 	check := strings.TrimSpace(instruction)
// 	if check == "" || strings.HasPrefix(instruction, "/") {
// 		return true
// 	}

// 	p.hasMoreLines = p.fileScanner.Scan()
// 	p.currentInstruction = p.fileScanner.Text()

// 	return p.hasMoreLines
// }


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
	// Get instruction type
	instType := instructionType(instruction)

	switch instType {

	// TODO: Handle symbols e.g. "@sum"
	case A_INSTRUCTION:
		addr := symbol(instruction)
		return DecimalStrToBinary15BitStr(addr)
	
	// TODO: Handle jump labels	e.g. "(LOOP)"
	case L_INSTRUCTION:
		return ""

	case C_INSTRUCTION: 
		var result strings.Builder

		result.WriteString("111")
		result.WriteString(GetCompBinary(comp(instruction)))
		result.WriteString(GetDestBinary(dest(instruction)))
		result.WriteString(GetJumpBinary(jump(instruction)))

		return result.String()

	default:
		panic("Unrecognized instruction type")
	}
}

func Parse(inputFile, outputFile *os.File) {
	fileScannerFirstPass := bufio.NewScanner(inputFile)
	fileScannerSecondPass := bufio.NewScanner(inputFile)
	fileScannerThirdPass := bufio.NewScanner(inputFile)

	symbolTable := make(map[string]string)

	// Initialize symbolTable: set R0=0, R1=1, ... R15=15
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
	
	// First pass: identify symbols and assign them RAM/ROM locations
	programCounter := 0

	for fileScannerFirstPass.Scan() {

		// Get ASM line
		line := strings.TrimSpace(fileScannerFirstPass.Text())
		if line == "" || strings.HasPrefix(line, "/") {
			continue
		}
		line = strings.Split(line, " ")[0]
		
		s := symbol(line)
		instType := instructionType(line)

		if instType == L_INSTRUCTION {
			symbolTable[s] = strconv.Itoa(programCounter)
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

	ramCounter := 16
	
	for fileScannerSecondPass.Scan() {

		// Get ASM line
		line := strings.TrimSpace(fileScannerSecondPass.Text())
		if line == "" || strings.HasPrefix(line, "/") {
			continue
		}
		line = strings.Split(line, " ")[0]
		
		s := symbol(line)
		instType := instructionType(line)

		if instType != A_INSTRUCTION {
			continue
		}

		if _, err := strconv.Atoi(s); err != nil && len(symbolTable[s]) == 0 {
			symbolTable[s] = strconv.Itoa(ramCounter)
			ramCounter++
		}
	}

	// Check for errors during scanning
	if err := fileScannerSecondPass.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input file: %v\n", err)
		os.Exit(1)
	}

	// Reset the file pointer to the beginning of the file
	_, err = inputFile.Seek(0, 0)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error seeking input file: %v\n", err)
		os.Exit(1)
	}
	
	// Second pass: translate assembly to binary
	for fileScannerThirdPass.Scan() {

		// Get ASM line
		line := strings.TrimSpace(fileScannerThirdPass.Text())
		if line == "" || strings.HasPrefix(line, "/") {
			continue
		}
		line = strings.Split(line, " ")[0]

		instType := instructionType(line)

		s := symbol(line)
		var nonSymbolizedLine string

		switch instType {
			case C_INSTRUCTION:				
				nonSymbolizedLine = line
			
			case L_INSTRUCTION:
				continue

			case A_INSTRUCTION:
				if _, err := strconv.Atoi(s); err != nil {
					nonSymbolizedLine = strings.Replace(line, s, symbolTable[s], 1)
				} else {
					nonSymbolizedLine = line
				}

			default:
				panic("Unrecognized instruction type")
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
	if err := fileScannerThirdPass.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input file: %v\n", err)
		os.Exit(1)
	}
}