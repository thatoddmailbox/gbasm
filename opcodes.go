package main

import (
	"log"
	"strconv"
	"strings"
)

type OperandType int

const (
	OperandConditionCode = iota
	OperandRegister8 = iota
	OperandRegister16 = iota
	OperandString = iota
	OperandValueIndirect = iota
	OperandValue = iota
)

type OpCodeInfo struct {
	ValidOperandCounts []int
}

var OpCodes_Table = map[string]OpCodeInfo{
	"ASCII": OpCodeInfo{[]int{1}},
	"ASCIZ": OpCodeInfo{[]int{1}},
	"DB": OpCodeInfo{[]int{-1}},
	"DW": OpCodeInfo{[]int{-1}},

	"ADD": OpCodeInfo{[]int{2}},
	"ADC": OpCodeInfo{[]int{2}},
	"SUB": OpCodeInfo{[]int{1}},
	"SBC": OpCodeInfo{[]int{2}},
	"AND": OpCodeInfo{[]int{1}},
	"XOR": OpCodeInfo{[]int{1}},
	"OR": OpCodeInfo{[]int{1}},
	"CP": OpCodeInfo{[]int{1}},

	"RLC": OpCodeInfo{[]int{1}},
	"RRC": OpCodeInfo{[]int{1}},
	"RL": OpCodeInfo{[]int{1}},
	"RR": OpCodeInfo{[]int{1}},
	"SLA": OpCodeInfo{[]int{1}},
	"SRA": OpCodeInfo{[]int{1}},
	"SWAP": OpCodeInfo{[]int{1}},
	"SRL": OpCodeInfo{[]int{1}},

	"BIT": OpCodeInfo{[]int{2}},
	"RES": OpCodeInfo{[]int{2}},
	"SET": OpCodeInfo{[]int{2}},
	"CALL": OpCodeInfo{[]int{1, 2}},
	"CPL": OpCodeInfo{[]int{0}},
	"JP": OpCodeInfo{[]int{1, 2}},
	"DEC": OpCodeInfo{[]int{1}},
	"INC": OpCodeInfo{[]int{1}},
	"DI": OpCodeInfo{[]int{0}},
	"EI": OpCodeInfo{[]int{0}},
	"HALT": OpCodeInfo{[]int{0}},
	"LD": OpCodeInfo{[]int{2}},
	"LDH": OpCodeInfo{[]int{2}},
	"LDI": OpCodeInfo{[]int{2}},
	"LDD": OpCodeInfo{[]int{2}},
	"NOP": OpCodeInfo{[]int{0}},
	"POP": OpCodeInfo{[]int{1}},
	"PUSH": OpCodeInfo{[]int{1}},
	"RET": OpCodeInfo{[]int{0, 1}},
	"RETI": OpCodeInfo{[]int{0}},
}

var OpCodes_Table_R = map[string]int {
	"B": 0,
	"C": 1,
	"D": 2,
	"E": 3,
	"H": 4,
	"L": 5,
	"[HL]": 6,
	"A": 7,
}

var OpCodes_Table_RP = map[string]int {
	"BC": 0,
	"DE": 1,
	"HL": 2,
	"SP": 3,
}

var OpCodes_Table_RP2 = map[string]int {
	"BC": 0,
	"DE": 1,
	"HL": 2,
	"AF": 3,
}

var OpCodes_Table_CC = map[string]int {
	"NZ": 0,
	"Z": 1,
	"NC": 2,
	"C": 3,
	"PO": 4,
	"PE": 5,
	"P": 6,
	"M": 7,
}

var OpCodes_Table_ALU = map[string]int {
	"ADD": 0,
	"ADC": 1,
	"SUB": 2,
	"SBC": 3,
	"AND": 4,
	"XOR": 5,
	"OR": 6,
	"CP": 7,
}

var OpCodes_Table_ROT = map[string]int {
	"RLC": 0,
	"RRC": 1,
	"RL": 2,
	"RR": 3,
	"SLA": 4,
	"SRA": 5,
	"SWAP": 6,
	"SRL": 7,
}

func OpCodes_GetOperandAsNumber(instruction Instruction, i int, fileBase string, lineNumber int) int {
	num, ok := Assembler_ParseNumber(instruction.Operands[i])
	if !ok {
		log.Fatalf("Expected number, got '%s' at %s:%d", instruction.Operands[i], fileBase, lineNumber)
	}
	return num
}

func OpCodes_GetOperandAsByte(instruction Instruction, i int, fileBase string, lineNumber int) byte {
	num := OpCodes_GetOperandAsNumber(instruction, i, fileBase, lineNumber)
	OpCodes_EnsureNumberIsByte(num, fileBase, lineNumber)
	return byte(num)
}

func OpCodes_GetOperandAsRegister8(instruction Instruction, i int, canBeIndirectHL bool, fileBase string, lineNumber int) string {
	foundType := OpCodes_GetOperandType(instruction, i, false)
	if foundType != OperandRegister8 {
		if !canBeIndirectHL || instruction.Operands[i] != "[HL]" {
			log.Fatalf("Expected 8-bit register, got '%s' at %s:%d", instruction.Operands[i], fileBase, lineNumber)
		}
	}
	return instruction.Operands[i]
}

func OpCodes_GetOperandAsRegister16(instruction Instruction, i int, fileBase string, lineNumber int) string {
	foundType := OpCodes_GetOperandType(instruction, i, false)
	if foundType != OperandRegister16 {
		log.Fatalf("Expected 16-bit register, got '%s' at %s:%d", instruction.Operands[i], fileBase, lineNumber)
	}
	return instruction.Operands[i]
}

func OpCodes_GetOperandAsString(instruction Instruction, i int, fileBase string, lineNumber int) string {
	foundType := OpCodes_GetOperandType(instruction, i, false)
	if foundType != OperandString {
		log.Fatalf("Expected string, got '%s' at %s:%d", instruction.Operands[i], fileBase, lineNumber)
	}
	return instruction.Operands[i][1:len(instruction.Operands[i]) - 1]
}

func OpCodes_GetOperandType(instruction Instruction, i int, canBeConditionCode bool) OperandType {
	operand := instruction.Operands[i]
	if canBeConditionCode && Utils_StringInSlice(operand, Parser_ConditionCodes) {
		return OperandConditionCode
	} else if Utils_StringInSlice(operand, Parser_8BitRegisterNames) || operand == "(HL)" {
		return OperandRegister8
	} else if Utils_StringInSlice(operand, Parser_16BitRegisterNames) {
		return OperandRegister16
	} else if operand[0] == '"' && operand[len(operand) - 1] == '"' {
		return OperandString
	} else {
		if strings.Contains(operand, "[") {
			return OperandValueIndirect
		}
		return OperandValue
	}
}

func OpCodes_EnsureNumberIsByte(num int, fileBase string, lineNumber int) {
	if num < 0 || num > 255 {
		log.Fatalf("Byte value %d out of range at %s:%d", num, fileBase, lineNumber)
	}
}

func OpCodes_AsmXZQP(x int, z int, q int, p int) byte {
	return byte((x << 6) | (p << 4) | (q << 3) | z)
}

func OpCodes_AsmXZY(x int, z int, y int) byte {
	return byte((x << 6) | (y << 3) | z)
}

func OpCodes_GetOutput(instruction Instruction, fileBase string, lineNumber int) []byte {
	info, ok := OpCodes_Table[instruction.Mnemonic]

	if !ok {
		log.Fatalf("Unknown instruction '%s' at %s:%d", instruction.Mnemonic, fileBase, lineNumber)
	}

	if info.ValidOperandCounts[0] != -1 && !Utils_IntInSlice(len(instruction.Operands), info.ValidOperandCounts) {
		log.Fatalf("Incorrect number of operands for instruction '%s' (got %d) at %s:%d", instruction.Mnemonic, len(instruction.Operands), fileBase, lineNumber)
	}

	var err error

	switch instruction.Mnemonic {
	case "ASCII":
		fallthrough
	case "ASCIZ":
		str := OpCodes_GetOperandAsString(instruction, 0, fileBase, lineNumber)
		output := []byte(str)
		if instruction.Mnemonic == "ASCIZ" {
			output = append(output, 0x00);
		}
		return output

	case "DB":
		// ok i guess technically it's not really an instruction but too bad
		output := []byte{}
		for i := 0; i < len(instruction.Operands); i++ {
			output = append(output, OpCodes_GetOperandAsByte(instruction, i, fileBase, lineNumber))
		}
		return output

	case "DW":
		// ok i guess technically this to isn't really an instruction but still too bad
		output := []byte{}
		for i := 0; i < len(instruction.Operands); i++ {
			num := OpCodes_GetOperandAsNumber(instruction, i, fileBase, lineNumber)
			output = append(output, byte(num & 0xFF))
			output = append(output, byte(num >> 8))
		}
		return output


	case "ADD":
		fallthrough
	case "ADC":
		fallthrough
	case "SUB":
		fallthrough
	case "SBC":
		fallthrough
	case "AND":
		fallthrough
	case "XOR":
		fallthrough
	case "OR":
		fallthrough
	case "CP":
		if len(instruction.Operands) == 2 {
			if instruction.Operands[0] != "A" {
				if instruction.Mnemonic != "ADD" || instruction.Operands[0] != "HL" {
					log.Fatalf("Invalid operand '%s' for %s at %s:%d", instruction.Operands[0], instruction.Mnemonic, fileBase, lineNumber)
				}
			}
		}

		if instruction.Mnemonic == "ADD" && instruction.Operands[0] == "HL" {
			// yay special case
			srcVal := OpCodes_GetOperandAsRegister16(instruction, 1, fileBase, lineNumber)
			return []byte{OpCodes_AsmXZQP(0, 1, 1, OpCodes_Table_RP[srcVal])}
		}

		targetIndex := len(instruction.Operands) - 1

		yVal := OpCodes_Table_ALU[instruction.Mnemonic]
		targetType := OpCodes_GetOperandType(instruction, targetIndex, false)
		if targetType == OperandValue {
			targetVal := OpCodes_GetOperandAsByte(instruction, targetIndex, fileBase, lineNumber)
			return []byte{OpCodes_AsmXZY(3, 6, yVal), targetVal}
		} else if (targetType == OperandRegister8 || instruction.Operands[targetIndex] == "[HL]") {
			targetVal := OpCodes_Table_R[instruction.Operands[targetIndex]]
			return []byte{OpCodes_AsmXZY(2, targetVal, yVal)}
		} else {
			log.Fatalf("Invalid operand '%s' for %s at %s:%d", instruction.Operands[targetIndex], instruction.Mnemonic, fileBase, lineNumber)
		}

	case "RLC":
		fallthrough
	case "RRC":
		fallthrough
	case "RL":
		fallthrough
	case "RR":
		fallthrough
	case "SLA":
		fallthrough
	case "SRA":
		fallthrough
	case "SWAP":
		fallthrough
	case "SRL":
		yVal := OpCodes_Table_ROT[instruction.Mnemonic]
		targetVal := OpCodes_Table_R[OpCodes_GetOperandAsRegister8(instruction, 0, true, fileBase, lineNumber)]
		return []byte{0xCB, OpCodes_AsmXZY(0, targetVal, yVal)}

	case "BIT":
		fallthrough
	case "RES":
		fallthrough
	case "SET":
		xVal := 1
		if instruction.Mnemonic == "RES" {
			xVal = 2
		} else if instruction.Mnemonic == "SET" {
			xVal = 3
		}
		target := OpCodes_GetOperandAsNumber(instruction, 0, fileBase, lineNumber)
		register := OpCodes_GetOperandAsRegister8(instruction, 1, true, fileBase, lineNumber)
		return []byte{0xCB, OpCodes_AsmXZY(xVal, OpCodes_Table_R[register], target)}

	case "CALL":
		fallthrough
	case "JP":
		operandCount := len(instruction.Operands)
		firstType := OpCodes_GetOperandType(instruction, 0, true)
		if operandCount == 1 {
			// direct jump
			if firstType == OperandValue {
				target := OpCodes_GetOperandAsNumber(instruction, 0, fileBase, lineNumber)
				firstByte := OpCodes_AsmXZY(3, 3, 0)
				if instruction.Mnemonic == "CALL" {
					firstByte = OpCodes_AsmXZQP(3, 5, 1, 0)
				}
				return []byte{firstByte, byte(target & 0xFF), byte(target >> 8)}
			} else if instruction.Mnemonic == "JP" && (instruction.Operands[0] == "HL" || instruction.Operands[0] == "[HL]") {
				return []byte{OpCodes_AsmXZQP(3, 1, 1, 2)}
			} else {
				log.Fatalf("Invalid operand '%s' for %s at %s:%d", instruction.Operands[0], instruction.Mnemonic, fileBase, lineNumber)
			}
		} else {
			// jump with condition code
			if firstType != OperandConditionCode {
				log.Fatalf("Invalid condition code '%s' for %s at %s:%d", instruction.Operands[0], instruction.Mnemonic, fileBase, lineNumber)
			}
			target := OpCodes_GetOperandAsNumber(instruction, 1, fileBase, lineNumber)
			z := 2
			if instruction.Mnemonic == "CALL" {
				z = 4
			}
			return []byte{OpCodes_AsmXZY(3, z, OpCodes_Table_CC[instruction.Operands[0]]), byte(target & 0xFF), byte(target >> 8)}
		}

	case "CPL":
		return []byte{0x2F}

	case "DEC":
		fallthrough
	case "INC":
		isINC := (instruction.Mnemonic == "INC")
		targetType := OpCodes_GetOperandType(instruction, 0, false)
		if (targetType == OperandRegister8 || instruction.Operands[0] == "[HL]") {
			targetVal := OpCodes_Table_R[instruction.Operands[0]]
			if isINC {
				return []byte{OpCodes_AsmXZY(0, 4, targetVal)}
			} else {
				return []byte{OpCodes_AsmXZY(0, 5, targetVal)}
			}
		} else if targetType == OperandRegister16 {
			targetVal := OpCodes_Table_RP[instruction.Operands[0]]
			if isINC {
				return []byte{OpCodes_AsmXZQP(0, 3, 0, targetVal)}
			} else {
				return []byte{OpCodes_AsmXZQP(0, 3, 1, targetVal)}
			}
		} else {
			log.Fatalf("Invalid operand '%s' for %s at %s:%d", instruction.Operands[0], instruction.Mnemonic, fileBase, lineNumber)
		}

	case "DI":
		return []byte{0xF3}

	case "EI":
		return []byte{0xFB}

	case "HALT":
		return []byte{0x76}

	case "LD":
		dstType := OpCodes_GetOperandType(instruction, 0, false)
		srcType := OpCodes_GetOperandType(instruction, 1, false)
		dstVal := 0
		srcVal := 0

		if dstType == OperandRegister8 {
			dstVal = OpCodes_Table_R[instruction.Operands[0]]
		} else if dstType == OperandRegister16 {
			dstVal = OpCodes_Table_RP[instruction.Operands[0]]
		} else {
			dstVal, err = strconv.Atoi(strings.Replace(strings.Replace(instruction.Operands[0], "[", "", -1), "]", "", -1))
			if err != nil { panic(err) }
		}

		if srcType == OperandRegister8 {
			srcVal = OpCodes_Table_R[instruction.Operands[1]]
		} else if srcType == OperandRegister16 {
			srcVal = OpCodes_Table_RP[instruction.Operands[1]]
		} else {
			srcVal, err = strconv.Atoi(strings.Replace(strings.Replace(instruction.Operands[1], "[", "", -1), "]", "", -1))
			if err != nil { panic(err) }
		}

		if (dstType == OperandRegister8 || instruction.Operands[0] == "[HL]") && srcType == OperandValue {
			if instruction.Operands[0] == "[HL]" {
				dstVal = OpCodes_Table_R["[HL]"]
			}
			OpCodes_EnsureNumberIsByte(srcVal, fileBase, lineNumber)
			return []byte{byte((dstVal << 3) | 6), byte(srcVal & 0xFF)}
		}
		if (dstType == OperandRegister8 || instruction.Operands[0] == "[HL]") && (srcType == OperandRegister8 || instruction.Operands[1] == "[HL]") {
			if instruction.Operands[0] == "[HL]" {
				dstVal = OpCodes_Table_R["[HL]"]
			}
			if instruction.Operands[1] == "[HL]" {
				srcVal = OpCodes_Table_R["[HL]"]
			}
			return []byte{byte(64 | (dstVal << 3) | srcVal)}
		}
		if dstType == OperandRegister16 && srcType == OperandValue {
			return []byte{byte((dstVal << 4) | 1), byte(srcVal & 0xFF), byte(srcVal >> 8)}
		}

		if dstType == OperandValueIndirect && instruction.Operands[1] == "A" {
			return []byte{0xEA, byte(dstVal & 0xFF), byte(dstVal >> 8)}
		}

		if instruction.Operands[0] == "A" && srcType == OperandValueIndirect {
			return []byte{0xFA, byte(srcVal & 0xFF), byte(srcVal >> 8)}
		}

		if instruction.Operands[0] == "[BC]" && instruction.Operands[1] == "A" { return []byte{OpCodes_AsmXZQP(0, 2, 0, 0)} }
		if instruction.Operands[0] == "[DE]" && instruction.Operands[1] == "A" { return []byte{OpCodes_AsmXZQP(0, 2, 0, 1)} }
		if instruction.Operands[0] == "A" && instruction.Operands[1] == "[BC]" { return []byte{OpCodes_AsmXZQP(0, 2, 1, 0)} }
		if instruction.Operands[0] == "A" && instruction.Operands[1] == "[DE]" { return []byte{OpCodes_AsmXZQP(0, 2, 1, 1)} }

		log.Fatalf("Invalid operands '%s' and '%s' for LD instruction at %s:%d", instruction.Operands[0], instruction.Operands[1], fileBase, lineNumber)
	
	case "LDH":
		dstType := OpCodes_GetOperandType(instruction, 0, false)
		srcType := OpCodes_GetOperandType(instruction, 1, false)
		srcVal := 0
		dstVal := 0

		if instruction.Operands[0] == "A" && srcType == OperandValueIndirect {
			srcVal, err = strconv.Atoi(strings.Replace(strings.Replace(instruction.Operands[1], "[", "", -1), "]", "", -1))
			if err != nil { panic(err) }

			if instruction.Operands[0] != "A" {
				log.Fatalf("LDH can only load into register A at %s:%d", instruction.Mnemonic, fileBase, lineNumber)
			}
			if srcVal >= 0xFF00 {
				srcVal = srcVal - 0xFF00
			}
			OpCodes_EnsureNumberIsByte(srcVal, fileBase, lineNumber)
			return []byte{0xF0, byte(srcVal & 0xFF)}
		} else if dstType == OperandValueIndirect && instruction.Operands[1] == "A" {
			dstVal, err = strconv.Atoi(strings.Replace(strings.Replace(instruction.Operands[0], "[", "", -1), "]", "", -1))
			if err != nil { panic(err) }

			if instruction.Operands[1] != "A" {
				log.Fatalf("LDH can only load from register A at %s:%d", instruction.Mnemonic, fileBase, lineNumber)
			}
			if dstVal >= 0xFF00 {
				dstVal = dstVal - 0xFF00
			}
			OpCodes_EnsureNumberIsByte(dstVal, fileBase, lineNumber)
			return []byte{0xE0, byte(dstVal & 0xFF)}
		} else {
			log.Fatalf("Invalid operands '%s' and '%s' for LDH instruction at %s:%d", instruction.Operands[0], instruction.Operands[1], fileBase, lineNumber)
		}

	case "LDI":
		fallthrough
	case "LDD":
		if instruction.Operands[0] == "A" && instruction.Operands[1] == "[HL]" {
			if instruction.Mnemonic == "LDI" {
				return []byte{0x2A}
			} else {
				return []byte{0x3A}
			}
		} else if instruction.Operands[0] == "[HL]" && instruction.Operands[1] == "A" {
			if instruction.Mnemonic == "LDI" {
				return []byte{0x22}
			} else {
				return []byte{0x32}
			}
		} else {
			log.Fatalf("Invalid operands '%s' and '%s' for %s instruction at %s:%d", instruction.Operands[0], instruction.Operands[1], instruction.Mnemonic, fileBase, lineNumber)
		}

	case "NOP":
		return []byte{0x00}

	case "POP":
		fallthrough
	case "PUSH":
		tableIndex, ok := OpCodes_Table_RP2[instruction.Operands[0]]
		if !ok {
			log.Fatalf("Invalid operand '%s' for %s instruction at %s:%d", instruction.Operands[0], instruction.Mnemonic, fileBase, lineNumber)
		}
		zVal := 1
		if instruction.Mnemonic == "PUSH" {
			zVal = 5
		}
		return []byte{OpCodes_AsmXZQP(3, zVal, 0, tableIndex)}

	case "RET":
		if len(instruction.Operands) == 0 {
			return []byte{OpCodes_AsmXZQP(3, 1, 1, 0)}
		} else {
			firstType := OpCodes_GetOperandType(instruction, 0, true)
			if firstType != OperandConditionCode {
				log.Fatalf("Invalid condition code '%s' for %s at %s:%d", instruction.Operands[0], instruction.Mnemonic, fileBase, lineNumber)
			}
			return []byte{OpCodes_AsmXZY(3, 0, OpCodes_Table_CC[instruction.Operands[0]])}
		}

	case "RETI":
		return []byte{0xD9}
	}
	return []byte{}
}