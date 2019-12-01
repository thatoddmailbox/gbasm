package main

import (
	"strconv"
	"strings"
	"testing"

	"github.com/thatoddmailbox/gbasm/utils"
)

func displayInstruction(instruction Instruction) string {
	return instruction.Mnemonic + " " + strings.Join(instruction.Operands, " ")
}

func prettyOutputArray(output []byte) string {
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

func tryTestInput(t *testing.T, instruction Instruction, expectedOutput []byte) {
	output := OpCodes_GetOutput(instruction, "test", 0)
	if !utils.ByteSlicesEqual(output, expectedOutput) {
		t.Errorf("Instruction '%s' assembled to %s, should have been %s", displayInstruction(instruction), prettyOutputArray(output), prettyOutputArray(expectedOutput))
	}
}

func TestControlInstructions(t *testing.T) {
	tryTestInput(t, Instruction{"CALL", []string{"1234"}}, []byte{0xCD, 0xD2, 0x04})
	tryTestInput(t, Instruction{"CALL", []string{"Z", "1234"}}, []byte{0xCC, 0xD2, 0x04})
	tryTestInput(t, Instruction{"JP", []string{"1234"}}, []byte{0xC3, 0xD2, 0x04})
	tryTestInput(t, Instruction{"JP", []string{"HL"}}, []byte{0xE9})
	tryTestInput(t, Instruction{"JP", []string{"[HL]"}}, []byte{0xE9})
	tryTestInput(t, Instruction{"JP", []string{"Z", "1234"}}, []byte{0xCA, 0xD2, 0x04})
	tryTestInput(t, Instruction{"RET", []string{}}, []byte{0xC9})
	tryTestInput(t, Instruction{"RET", []string{"Z"}}, []byte{0xC8})
	tryTestInput(t, Instruction{"RETI", []string{}}, []byte{0xD9})
}

func TestBitInstructions(t *testing.T) {
	tryTestInput(t, Instruction{"BIT", []string{"1", "A"}}, []byte{0xCB, 0x4F})
	tryTestInput(t, Instruction{"BIT", []string{"2", "B"}}, []byte{0xCB, 0x50})
	tryTestInput(t, Instruction{"BIT", []string{"3", "[HL]"}}, []byte{0xCB, 0x5E})

	tryTestInput(t, Instruction{"RES", []string{"1", "A"}}, []byte{0xCB, 0x8F})
	tryTestInput(t, Instruction{"RES", []string{"2", "B"}}, []byte{0xCB, 0x90})
	tryTestInput(t, Instruction{"RES", []string{"3", "[HL]"}}, []byte{0xCB, 0x9E})

	tryTestInput(t, Instruction{"SET", []string{"1", "A"}}, []byte{0xCB, 0xCF})
	tryTestInput(t, Instruction{"SET", []string{"2", "B"}}, []byte{0xCB, 0xD0})
	tryTestInput(t, Instruction{"SET", []string{"3", "[HL]"}}, []byte{0xCB, 0xDE})
}

func TestLoadInstructions(t *testing.T) {
	tryTestInput(t, Instruction{"LD", []string{"A", "A"}}, []byte{0x7f})
	tryTestInput(t, Instruction{"LD", []string{"A", "B"}}, []byte{0x78})
	tryTestInput(t, Instruction{"LD", []string{"A", "C"}}, []byte{0x79})
	tryTestInput(t, Instruction{"LD", []string{"A", "D"}}, []byte{0x7a})
	tryTestInput(t, Instruction{"LD", []string{"A", "E"}}, []byte{0x7b})
	tryTestInput(t, Instruction{"LD", []string{"A", "H"}}, []byte{0x7c})
	tryTestInput(t, Instruction{"LD", []string{"A", "L"}}, []byte{0x7d})
	tryTestInput(t, Instruction{"LD", []string{"B", "66"}}, []byte{0x06, 0x42})
	tryTestInput(t, Instruction{"LD", []string{"BC", "1234"}}, []byte{0x01, 0xD2, 0x04})
	tryTestInput(t, Instruction{"LD", []string{"[HL]", "66"}}, []byte{0x36, 0x42})
	tryTestInput(t, Instruction{"LD", []string{"[BC]", "A"}}, []byte{0x02})
	tryTestInput(t, Instruction{"LD", []string{"[DE]", "A"}}, []byte{0x12})
	tryTestInput(t, Instruction{"LD", []string{"[HL]", "A"}}, []byte{0x77})
	tryTestInput(t, Instruction{"LD", []string{"A", "[HL]"}}, []byte{0x7E})
	tryTestInput(t, Instruction{"LD", []string{"[1234]", "A"}}, []byte{0xEA, 0xD2, 0x04})
	tryTestInput(t, Instruction{"LD", []string{"A", "[1234]"}}, []byte{0xFA, 0xD2, 0x04})
	tryTestInput(t, Instruction{"LDH", []string{"A", "[40]"}}, []byte{0xF0, 0x28})
	tryTestInput(t, Instruction{"LDH", []string{"A", "[65320]"}}, []byte{0xF0, 0x28})
	tryTestInput(t, Instruction{"LDH", []string{"[40]", "A"}}, []byte{0xE0, 0x28})
	tryTestInput(t, Instruction{"LDH", []string{"[65320]", "A"}}, []byte{0xE0, 0x28})
	tryTestInput(t, Instruction{"LDI", []string{"[HL]", "A"}}, []byte{0x22})
	tryTestInput(t, Instruction{"LDI", []string{"A", "[HL]"}}, []byte{0x2A})
	tryTestInput(t, Instruction{"LDD", []string{"[HL]", "A"}}, []byte{0x32})
	tryTestInput(t, Instruction{"LDD", []string{"A", "[HL]"}}, []byte{0x3A})
}

func TestALUInstructions(t *testing.T) {
	tryTestInput(t, Instruction{"ADD", []string{"A", "66"}}, []byte{0xC6, 0x42})
	tryTestInput(t, Instruction{"ADD", []string{"A", "B"}}, []byte{0x80})
	tryTestInput(t, Instruction{"ADD", []string{"A", "[HL]"}}, []byte{0x86})

	tryTestInput(t, Instruction{"ADD", []string{"HL", "BC"}}, []byte{0x09})
	tryTestInput(t, Instruction{"ADD", []string{"HL", "DE"}}, []byte{0x19})
	tryTestInput(t, Instruction{"ADD", []string{"HL", "HL"}}, []byte{0x29})
	tryTestInput(t, Instruction{"ADD", []string{"HL", "SP"}}, []byte{0x39})

	tryTestInput(t, Instruction{"ADC", []string{"A", "66"}}, []byte{0xCE, 0x42})
	tryTestInput(t, Instruction{"ADC", []string{"A", "B"}}, []byte{0x88})
	tryTestInput(t, Instruction{"ADC", []string{"A", "[HL]"}}, []byte{0x8E})

	tryTestInput(t, Instruction{"SUB", []string{"66"}}, []byte{0xD6, 0x42})
	tryTestInput(t, Instruction{"SUB", []string{"B"}}, []byte{0x90})
	tryTestInput(t, Instruction{"SUB", []string{"[HL]"}}, []byte{0x96})

	tryTestInput(t, Instruction{"SBC", []string{"A", "66"}}, []byte{0xDE, 0x42})
	tryTestInput(t, Instruction{"SBC", []string{"A", "B"}}, []byte{0x98})
	tryTestInput(t, Instruction{"SBC", []string{"A", "[HL]"}}, []byte{0x9E})

	tryTestInput(t, Instruction{"AND", []string{"66"}}, []byte{0xE6, 0x42})
	tryTestInput(t, Instruction{"AND", []string{"B"}}, []byte{0xA0})
	tryTestInput(t, Instruction{"AND", []string{"[HL]"}}, []byte{0xA6})

	tryTestInput(t, Instruction{"XOR", []string{"66"}}, []byte{0xEE, 0x42})
	tryTestInput(t, Instruction{"XOR", []string{"B"}}, []byte{0xA8})
	tryTestInput(t, Instruction{"XOR", []string{"[HL]"}}, []byte{0xAE})

	tryTestInput(t, Instruction{"OR", []string{"66"}}, []byte{0xF6, 0x42})
	tryTestInput(t, Instruction{"OR", []string{"B"}}, []byte{0xB0})
	tryTestInput(t, Instruction{"OR", []string{"[HL]"}}, []byte{0xB6})

	tryTestInput(t, Instruction{"CP", []string{"66"}}, []byte{0xFE, 0x42})
	tryTestInput(t, Instruction{"CP", []string{"A"}}, []byte{0xBF})
	tryTestInput(t, Instruction{"CP", []string{"[HL]"}}, []byte{0xBE})
}

func TestMathInstructions(t *testing.T) {
	tryTestInput(t, Instruction{"DEC", []string{"A"}}, []byte{0x3D})
	tryTestInput(t, Instruction{"DEC", []string{"HL"}}, []byte{0x2B})
	tryTestInput(t, Instruction{"INC", []string{"A"}}, []byte{0x3C})
	tryTestInput(t, Instruction{"INC", []string{"BC"}}, []byte{0x03})
	tryTestInput(t, Instruction{"INC", []string{"HL"}}, []byte{0x23})
	tryTestInput(t, Instruction{"INC", []string{"[HL]"}}, []byte{0x34})
}

func TestROTInstructions(t *testing.T) {
	tryTestInput(t, Instruction{"CPL", []string{}}, []byte{0x2F})

	tryTestInput(t, Instruction{"RLC", []string{"A"}}, []byte{0xCB, 0x07})
	tryTestInput(t, Instruction{"RLC", []string{"B"}}, []byte{0xCB, 0x00})
	tryTestInput(t, Instruction{"RLC", []string{"[HL]"}}, []byte{0xCB, 0x06})

	tryTestInput(t, Instruction{"RRC", []string{"A"}}, []byte{0xCB, 0x0F})
	tryTestInput(t, Instruction{"RRC", []string{"B"}}, []byte{0xCB, 0x08})
	tryTestInput(t, Instruction{"RRC", []string{"[HL]"}}, []byte{0xCB, 0x0E})

	tryTestInput(t, Instruction{"RL", []string{"A"}}, []byte{0xCB, 0x17})
	tryTestInput(t, Instruction{"RL", []string{"B"}}, []byte{0xCB, 0x10})
	tryTestInput(t, Instruction{"RL", []string{"[HL]"}}, []byte{0xCB, 0x16})

	tryTestInput(t, Instruction{"RR", []string{"A"}}, []byte{0xCB, 0x1F})
	tryTestInput(t, Instruction{"RR", []string{"B"}}, []byte{0xCB, 0x18})
	tryTestInput(t, Instruction{"RR", []string{"[HL]"}}, []byte{0xCB, 0x1E})

	tryTestInput(t, Instruction{"SLA", []string{"A"}}, []byte{0xCB, 0x27})
	tryTestInput(t, Instruction{"SLA", []string{"B"}}, []byte{0xCB, 0x20})
	tryTestInput(t, Instruction{"SLA", []string{"[HL]"}}, []byte{0xCB, 0x26})

	tryTestInput(t, Instruction{"SRA", []string{"A"}}, []byte{0xCB, 0x2F})
	tryTestInput(t, Instruction{"SRA", []string{"B"}}, []byte{0xCB, 0x28})
	tryTestInput(t, Instruction{"SRA", []string{"[HL]"}}, []byte{0xCB, 0x2E})

	tryTestInput(t, Instruction{"SWAP", []string{"A"}}, []byte{0xCB, 0x37})
	tryTestInput(t, Instruction{"SWAP", []string{"B"}}, []byte{0xCB, 0x30})
	tryTestInput(t, Instruction{"SWAP", []string{"[HL]"}}, []byte{0xCB, 0x36})

	tryTestInput(t, Instruction{"SRL", []string{"A"}}, []byte{0xCB, 0x3F})
	tryTestInput(t, Instruction{"SRL", []string{"B"}}, []byte{0xCB, 0x38})
	tryTestInput(t, Instruction{"SRL", []string{"[HL]"}}, []byte{0xCB, 0x3E})
}

func TestMiscInstructions(t *testing.T) {
	tryTestInput(t, Instruction{"DB", []string{"66"}}, []byte{0x42})
	tryTestInput(t, Instruction{"DB", []string{"66", "66"}}, []byte{0x42, 0x42})
	tryTestInput(t, Instruction{"DB", []string{"66", "66", "66"}}, []byte{0x42, 0x42, 0x42})
	tryTestInput(t, Instruction{"DW", []string{"1234"}}, []byte{0xD2, 0x04})

	tryTestInput(t, Instruction{"ASCII", []string{"\"hello\""}}, []byte{0x68, 0x65, 0x6C, 0x6C, 0x6F})
	tryTestInput(t, Instruction{"ASCIZ", []string{"\"hello\""}}, []byte{0x68, 0x65, 0x6C, 0x6C, 0x6F, 0x00})

	tryTestInput(t, Instruction{"DI", []string{}}, []byte{0xF3})
	tryTestInput(t, Instruction{"EI", []string{}}, []byte{0xFB})
	tryTestInput(t, Instruction{"HALT", []string{}}, []byte{0x76})
	tryTestInput(t, Instruction{"NOP", []string{}}, []byte{0x00})
	tryTestInput(t, Instruction{"PUSH", []string{"BC"}}, []byte{0xC5})
	tryTestInput(t, Instruction{"PUSH", []string{"DE"}}, []byte{0xD5})
	tryTestInput(t, Instruction{"PUSH", []string{"HL"}}, []byte{0xE5})
	tryTestInput(t, Instruction{"PUSH", []string{"AF"}}, []byte{0xF5})
	tryTestInput(t, Instruction{"POP", []string{"BC"}}, []byte{0xC1})
	tryTestInput(t, Instruction{"POP", []string{"DE"}}, []byte{0xD1})
	tryTestInput(t, Instruction{"POP", []string{"HL"}}, []byte{0xE1})
	tryTestInput(t, Instruction{"POP", []string{"AF"}}, []byte{0xF1})
}
