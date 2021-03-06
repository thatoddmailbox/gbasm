package main

import (
	"bufio"
	"log"
	"os"
	"path"
	"strings"

	"github.com/thatoddmailbox/gbasm/parser"
	"github.com/thatoddmailbox/gbasm/rom"
	"github.com/thatoddmailbox/gbasm/utils"
)

type Instruction struct {
	Mnemonic string
	Operands []string
}

func Assembler_ParseFile(filePath string, origin int, maxLength int) int {
	fileBase := path.Base(filePath)

	log.Printf("Parsing file %s...\n", fileBase)

	Assembler_FindLabelsInFile(filePath, fileBase)
	Assembler_ParseFilePass(filePath, fileBase, origin, maxLength, 0) // first pass just finds what labels are pointing to
	return Assembler_ParseFilePass(filePath, fileBase, origin, maxLength, 1)
}

func Assembler_FindLabelsInFile(filePath string, fileBase string) {
	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())
		if len(line) == 0 {
			continue
		}
		if line[0] == '.' {
			// is it an include?
			if strings.HasPrefix(line, ".incasm") {
				// if so, get those labels
				instructionParts := strings.Split(line, " ")
				includedFilePath := path.Join(path.Dir(filePath), strings.Replace(instructionParts[1], "\"", "", -1))
				Assembler_FindLabelsInFile(includedFilePath, path.Base(includedFilePath))
			}
		}
		if line[len(line)-1] == ':' {
			// it's a label
			labelName := line[:len(line)-1]

			_, existsInDefs := rom.Current.Definitions[labelName]
			existsInUnpointedDefs := utils.StringInSlice(labelName, rom.Current.UnpointedDefinitions)
			if existsInDefs || existsInUnpointedDefs {
				log.Fatalf("Tried to declare already existing label or constant '%s' at %s:%d", labelName, fileBase, lineNumber)
			}

			rom.Current.UnpointedDefinitions = append(rom.Current.UnpointedDefinitions, labelName)
		}
	}
}

func Assembler_ParseFilePass(filePath string, fileBase string, origin int, maxLength int, pass int) int {
	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	outputIndex := origin

	scanner := bufio.NewScanner(file)
	lineNumber := 0
	inMultilineComment := false
	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())
		if len(line) == 0 {
			continue
		}

		if inMultilineComment {
			if len(line) > 1 && line[len(line)-2] == '*' && line[len(line)-1] == '/' {
				inMultilineComment = false
			}
			continue
		}

		if line[0] == '.' && len(line) > 1 {
			// special instruction
			instructionParts := strings.Split(line, " ")
			switch instructionParts[0][1:] {
			case "def":
				if pass != 0 {
					// defines only apply on the first pass
					continue
				}
				key := instructionParts[1]

				val, valid := parser.ParseNumber(parser.SimplifyPotentialExpression(instructionParts[2], pass, fileBase, lineNumber))
				if !valid {
					log.Fatalf("Expected number, got '%s' at %s:%d", instructionParts[2], fileBase, lineNumber)
				}

				_, exists := rom.Current.Definitions[key]
				if exists {
					log.Fatalf("Tried to declare already existing label or constant '%s' at %s:%d", key, fileBase, lineNumber)
				}

				rom.Current.Definitions[key] = val

			case "org":
				newOrigin, valid := parser.ParseNumber(instructionParts[1])
				if !valid {
					log.Fatalf("Expected number, got '%s' at %s:%d", instructionParts[1], fileBase, lineNumber)
				}
				outputIndex = newOrigin

			case "incasm":
				includedFilePath := path.Join(path.Dir(filePath), strings.Replace(instructionParts[1], "\"", "", -1))
				outputIndex = Assembler_ParseFilePass(includedFilePath, path.Base(includedFilePath), outputIndex, maxLength, pass)

			default:
				log.Fatalf("Unknown special instruction '%s' at %s:%d", instructionParts[0][1:], fileBase, lineNumber)
			}
		} else {
			// it's either a comment, label, or instruction

			// is it a label?
			if line[len(line)-1] == ':' {
				// it is
				if pass != 0 {
					// labels only apply on the first pass
					continue
				}

				labelName := line[:len(line)-1]

				_, exists := rom.Current.Definitions[labelName]
				if exists {
					log.Fatalf("Tried to declare already existing label or constant '%s' at %s:%d", labelName, fileBase, lineNumber)
				}

				rom.Current.Definitions[labelName] = outputIndex
			} else {
				// parse it character-by-character
				buf := ""
				instruction := Instruction{}
				foundAnInstruction := false
				inAString := false

				for i := 0; i < len(line); i++ {
					char := line[i]
					if !inAString && ((char == '/' && (len(line)-i) > 1) || char == ';') {
						if char == ';' || line[i+1] == '/' {
							// single-line comment
							break
						} else if line[i+1] == '*' {
							// start of multi-line comment
							inMultilineComment = true
							break
						} else {
							log.Fatalf("Unexpected '%s' after '/' at %s:%d", string(line[i+1]), fileBase, lineNumber)
						}
					} else if char == ' ' && instruction.Mnemonic == "" {
						// yay we have a mnemonic
						instruction.Mnemonic = strings.ToUpper(buf)
						foundAnInstruction = true
						buf = ""
					} else if char == '"' {
						inAString = !inAString
						buf += string(char)
					} else if char == ',' && foundAnInstruction && !inAString {
						// yay we have an operand
						instruction.Operands = append(instruction.Operands, buf)
						foundAnInstruction = true
						buf = ""
					} else {
						buf += string(char)
					}
				}
				if !foundAnInstruction && buf != "" {
					// it's an instruction with no operands
					foundAnInstruction = true
					instruction.Mnemonic = strings.ToUpper(buf)
					buf = ""
				}
				if foundAnInstruction {
					if buf != "" {
						// add any extra as the last operand
						instruction.Operands = append(instruction.Operands, buf)
					}

					// process the operands and any expressions in them
					for i := 0; i < len(instruction.Operands); i++ {
						instruction.Operands[i] = parser.SimplifyPotentialExpression(instruction.Operands[i], pass, fileBase, lineNumber)

						if utils.StringInSlice(strings.ToUpper(instruction.Operands[i]), append(append(parser.RegisterNames8, parser.RegisterNames16...), parser.ConditionCodes...)) {
							// capitalize register and condition code names
							instruction.Operands[i] = strings.ToUpper(instruction.Operands[i])
						}
					}

					// now, actually assemble the instruction
					outputIndex = Assembler_AssembleInstruction(instruction, outputIndex, fileBase, lineNumber)
				}
			}
		}
	}

	if err = scanner.Err(); err != nil {
		panic(err)
	}

	return outputIndex
}

func Assembler_AssembleInstruction(instruction Instruction, outputIndex int, fileBase string, lineNumber int) int {
	output := OpCodes_GetOutput(instruction, fileBase, lineNumber)
	for i := 0; i < len(output); i++ {
		rom.Current.Output[outputIndex] = output[i]
		outputIndex++
	}
	rom.Current.UsedByteCount += len(output)
	return outputIndex
}
