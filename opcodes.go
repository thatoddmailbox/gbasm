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
	OperandValueIndirect = iota
	OperandValue = iota
)

type OpCodeInfo struct {
	ValidOperandCounts []int
}

var OpCodes_Table = map[string]OpCodeInfo{
	"DB": OpCodeInfo{[]int{1}},

	"BIT": OpCodeInfo{[]int{2}},
	"CALL": OpCodeInfo{[]int{1, 2}},
	"JP": OpCodeInfo{[]int{1, 2}},
	"DEC": OpCodeInfo{[]int{1}},
	"INC": OpCodeInfo{[]int{1}},
	"LD": OpCodeInfo{[]int{2}},
	"NOP": OpCodeInfo{[]int{0}},
	"RET": OpCodeInfo{[]int{0, 1}},
}

var OpCodes_Table_R = map[string]int {
	"B": 0,
	"C": 1,
	"D": 2,
	"E": 3,
	"H": 4,
	"L": 5,
	"(HL)": 6,
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

func OpCodes_GetOperandAsNumber(instruction Instruction, i int, fileBase string, lineNumber int) int {
	num, ok := Assembler_ParseNumber(instruction.Operands[i])
	if !ok {
		log.Fatalf("Expected number, got '%s' at %s:%d", instruction.Operands[i], fileBase, lineNumber)
	}
	return num
}

func OpCodes_GetOperandAsByte(instruction Instruction, i int, fileBase string, lineNumber int) byte {
	num := OpCodes_GetOperandAsNumber(instruction, i, fileBase, lineNumber)
	if num < 0 || num > 255 {
		log.Fatalf("Byte value %d out of range at %s:%d", num, fileBase, lineNumber)
	}
	return byte(num)
}


func OpCodes_GetOperandAsRegister8(instruction Instruction, i int, fileBase string, lineNumber int) string {
	foundType := OpCodes_GetOperandType(instruction, i, false)
	if foundType != OperandRegister8 {
		log.Fatalf("Expected register, got '%s' at %s:%d", instruction.Operands[i], fileBase, lineNumber)
	}
	return instruction.Operands[i]
}


func OpCodes_GetOperandType(instruction Instruction, i int, canBeConditionCode bool) OperandType {
	operand := instruction.Operands[i]
	if canBeConditionCode && Utils_StringInSlice(operand, Parser_ConditionCodes) {
		return OperandConditionCode
	} else if Utils_StringInSlice(operand, Parser_8BitRegisterNames) || operand == "(HL)" {
		return OperandRegister8
	} else if Utils_StringInSlice(operand, Parser_16BitRegisterNames) {
		return OperandRegister16
	} else {
		if strings.Contains(operand, "(") {
			return OperandValueIndirect
		}
		return OperandValue
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

	if !Utils_IntInSlice(len(instruction.Operands), info.ValidOperandCounts) {
		log.Fatalf("Incorrect number of operands for instruction '%s' (got %d) at %s:%d", instruction.Mnemonic, len(instruction.Operands), fileBase, lineNumber)
	}

	var err error

	switch instruction.Mnemonic {
	case "DB":
		// ok i guess technically it's not really an instruction but too bad
		return []byte{OpCodes_GetOperandAsByte(instruction, 0, fileBase, lineNumber)}

	case "BIT":
		target := OpCodes_GetOperandAsNumber(instruction, 0, fileBase, lineNumber)
		register := OpCodes_GetOperandAsRegister8(instruction, 1, fileBase, lineNumber)
		return []byte{0xCB, OpCodes_AsmXZY(1, OpCodes_Table_R[register], target)}

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
			} else if instruction.Mnemonic == "JP" && (instruction.Operands[0] == "HL" || instruction.Operands[0] == "(HL)") {
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

	case "DEC":
		fallthrough
	case "INC":
		isINC := (instruction.Mnemonic == "INC")
		targetType := OpCodes_GetOperandType(instruction, 0, false)
		if targetType == OperandRegister8 {
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
			dstVal, err = strconv.Atoi(strings.Replace(strings.Replace(instruction.Operands[0], "(", "", 1), ")", "", 1))
			if err != nil { panic(err) }
		}

		if srcType == OperandRegister8 {
			srcVal = OpCodes_Table_R[instruction.Operands[1]]
		} else if srcType == OperandRegister16 {
			srcVal = OpCodes_Table_RP[instruction.Operands[1]]
		} else {
			srcVal, err = strconv.Atoi(strings.Replace(strings.Replace(instruction.Operands[1], "(", "", 1), ")", "", 1))
			if err != nil { panic(err) }
		}

		if dstType == OperandRegister16 && srcType == OperandValue {
			return []byte{byte((dstVal << 4) | 1), byte(srcVal & 0xFF), byte(srcVal >> 8)}
		}
		if dstType == OperandRegister8 && srcType == OperandValue {
			return []byte{byte((dstVal << 3) | 6), byte(srcVal & 0xFF)}
		}
		if dstType == OperandRegister8 && srcType == OperandRegister8 {
			return []byte{byte(64 | (dstVal << 3) | srcVal)}
		}

		if dstType == OperandValueIndirect && instruction.Operands[1] == "HL" {
			return []byte{OpCodes_AsmXZQP(0, 2, 0, 2), byte(dstVal & 0xFF), byte(dstVal >> 8)}
		}
		if dstType == OperandValueIndirect && instruction.Operands[1] == "A" {
			return []byte{OpCodes_AsmXZQP(0, 2, 0, 3), byte(dstVal & 0xFF), byte(dstVal >> 8)}
		}

		if instruction.Operands[0] == "HL" && srcType == OperandValueIndirect {
			return []byte{OpCodes_AsmXZQP(0, 2, 1, 2), byte(srcVal & 0xFF), byte(srcVal >> 8)}
		}
		if instruction.Operands[0] == "A" && srcType == OperandValueIndirect {
			return []byte{OpCodes_AsmXZQP(0, 2, 1, 3), byte(srcVal & 0xFF), byte(srcVal >> 8)}
		}

		if instruction.Operands[0] == "(BC)" && instruction.Operands[1] == "A" { return []byte{OpCodes_AsmXZQP(0, 2, 0, 0)} }
		if instruction.Operands[0] == "(DE)" && instruction.Operands[1] == "A" { return []byte{OpCodes_AsmXZQP(0, 2, 0, 1)} }
		if instruction.Operands[0] == "A" && instruction.Operands[1] == "(BC)" { return []byte{OpCodes_AsmXZQP(0, 2, 1, 0)} }
		if instruction.Operands[0] == "A" && instruction.Operands[1] == "(DE)" { return []byte{OpCodes_AsmXZQP(0, 2, 1, 1)} }

		log.Fatalf("Invalid operands '%s' and '%s' for LD instruction at %s:%d", instruction.Operands[0], instruction.Operands[1], fileBase, lineNumber)
		
	case "NOP":
		return []byte{0x00}

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
	}
	return []byte{}
}