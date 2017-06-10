package main

import (
	"strconv"
	"strings"
	"testing"
)

func DisplayInstruction(instruction Instruction) string {
	return instruction.Mnemonic + " " + strings.Join(instruction.Operands, " ")
}

func PrettyOutputArray(output []byte) string {
	outStr := "["
	for i, c := range output {
		if i != 0 {
			outStr += ", "
		}
		outStr += strconv.Itoa(int(c))
	}
	outStr += "]"
	return outStr
}

func TryTestInput(t *testing.T, instruction Instruction, expectedOutput []byte) {
	output := OpCodes_GetOutput(instruction, "test", 0)
	if !Utils_ByteSlicesEqual(output, expectedOutput) {
		t.Errorf("Instruction '%s' assembled to %s, should have been %s", DisplayInstruction(instruction), PrettyOutputArray(output), PrettyOutputArray(expectedOutput))
	}
}

func TestControlInstructions(t *testing.T) {
	TryTestInput(t, Instruction{"CALL", []string{"1234"}}, []byte{0xCD, 0xD2, 0x04})
	TryTestInput(t, Instruction{"CALL", []string{"Z", "1234"}}, []byte{0xCC, 0xD2, 0x04})
	TryTestInput(t, Instruction{"JP", []string{"1234"}}, []byte{0xC3, 0xD2, 0x04})
	TryTestInput(t, Instruction{"JP", []string{"HL"}}, []byte{0xE9})
	TryTestInput(t, Instruction{"JP", []string{"(HL)"}}, []byte{0xE9})
	TryTestInput(t, Instruction{"JP", []string{"Z", "1234"}}, []byte{0xCA, 0xD2, 0x04})
	TryTestInput(t, Instruction{"RET", []string{}}, []byte{0xC9})
	TryTestInput(t, Instruction{"RET", []string{"Z"}}, []byte{0xC8})
}

func TestCompareInstructions(t *testing.T) {
	TryTestInput(t, Instruction{"BIT", []string{"1", "A"}}, []byte{0xCB, 0x4F})
	TryTestInput(t, Instruction{"BIT", []string{"2", "B"}}, []byte{0xCB, 0x50})
}

func TestLoadInstructions(t *testing.T) {
	TryTestInput(t, Instruction{"LD", []string{"A", "A"}}, []byte{0x7f})
	TryTestInput(t, Instruction{"LD", []string{"A", "B"}}, []byte{0x78})
	TryTestInput(t, Instruction{"LD", []string{"A", "C"}}, []byte{0x79})
	TryTestInput(t, Instruction{"LD", []string{"A", "D"}}, []byte{0x7a})
	TryTestInput(t, Instruction{"LD", []string{"A", "E"}}, []byte{0x7b})
	TryTestInput(t, Instruction{"LD", []string{"A", "H"}}, []byte{0x7c})
	TryTestInput(t, Instruction{"LD", []string{"A", "L"}}, []byte{0x7d})
	TryTestInput(t, Instruction{"LD", []string{"B", "66"}}, []byte{0x06, 0x42})
	TryTestInput(t, Instruction{"LD", []string{"BC", "1234"}}, []byte{0x01, 0xD2, 0x04})
	TryTestInput(t, Instruction{"LD", []string{"(BC)", "A"}}, []byte{0x02})
	TryTestInput(t, Instruction{"LD", []string{"(DE)", "A"}}, []byte{0x12})
	TryTestInput(t, Instruction{"LD", []string{"(HL)", "A"}}, []byte{0x77})
	TryTestInput(t, Instruction{"LD", []string{"(1234)", "HL"}}, []byte{0x22, 0xD2, 0x04})
	TryTestInput(t, Instruction{"LD", []string{"(1234)", "A"}}, []byte{0x32, 0xD2, 0x04})
	TryTestInput(t, Instruction{"LD", []string{"HL", "(1234)"}}, []byte{0x2A, 0xD2, 0x04})
	TryTestInput(t, Instruction{"LD", []string{"A", "(1234)"}}, []byte{0x3A, 0xD2, 0x04})
}

func TestMathInstructions(t *testing.T) {
	TryTestInput(t, Instruction{"DEC", []string{"A"}}, []byte{0x3D})
	TryTestInput(t, Instruction{"DEC", []string{"HL"}}, []byte{0x2B})
	TryTestInput(t, Instruction{"INC", []string{"A"}}, []byte{0x3C})
	TryTestInput(t, Instruction{"INC", []string{"HL"}}, []byte{0x23})
}

func TestMiscInstructions(t *testing.T) {
	TryTestInput(t, Instruction{"NOP", []string{}}, []byte{0x00})
}