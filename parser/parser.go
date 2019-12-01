package parser

import (
	"log"
	"strconv"
	"strings"

	"github.com/thatoddmailbox/gbasm/rom"
	"github.com/thatoddmailbox/gbasm/utils"
)

var RegisterNames8 = []string{
	"A",
	"B",
	"C",
	"D",
	"E",
	"F",
	"H",
	"L",
}

var RegisterNames16 = []string{
	"AF",
	"BC",
	"[BC]",
	"DE",
	"[DE]",
	"HL",
	"[HL]",
	"PC",
	"SP",
}

var ConditionCodes = []string{
	"Z",
	"NZ",
	"NC",
	"C",
	"PO",
	"PE",
	"P",
	"M",
}

var Tokens = map[string]int{
	"*":  6,
	"/":  6,
	"+":  5,
	"-":  5,
	">>": 4,
	"<<": 4,
	"&":  3,
	"^":  2,
	"|":  1,
}

// ParseNumber parses the given string as a numeric constant.
func ParseNumber(numString string) (int, bool) {
	numString = strings.TrimSpace(numString)
	if strings.HasPrefix(numString, "0b") {
		// it's a binary number
		num, err := strconv.ParseInt(numString[2:], 2, 0)
		if err != nil {
			return 0, false
		}
		return int(num), true
	}
	if len(numString) == 3 && numString[0] == '\'' && numString[2] == '\'' {
		// it's a character
		return int(numString[1]), true
	}
	num, err := strconv.ParseInt(numString, 0, 0)
	if err != nil {
		return 0, false
	}
	return int(num), true
}

func ParseExpressionTokens(expression string) []string {
	tokens := []string{}
	buf := ""
	inAString := false

	for i := 0; i < len(expression); i++ {
		char := expression[i]
		if char == '"' || char == '\'' {
			buf = buf + string(char)
			inAString = !inAString
		} else if inAString {
			buf = buf + string(char)
		} else if (char == '(' || char == ')') ||
			(char == '+' || char == '-' || char == '*' || char == '/') ||
			(char == '&' || char == '|' || char == '^') ||
			(char == '>' || char == '<') {
			if buf != "" {
				tokens = append(tokens, buf)
			}
			buf = ""

			if char == '>' || char == '<' {
				// is it a bitshift operator?
				// if so, is the next character the same?
				if i+1 < len(expression) && char == expression[i+1] {
					// then add the token and ignore the next character
					tokens = append(tokens, string(char)+string(char))
					i++
					continue
				}
			}

			tokens = append(tokens, string(char))
		} else if char == ' ' {
			if buf != "" {
				tokens = append(tokens, buf)
			}
			buf = ""
		} else {
			buf = buf + string(char)
		}
	}

	if buf != "" {
		tokens = append(tokens, buf)
	}

	return tokens
}

func IsSecondOperatorMoreImportantThanFirst(first string, second string) bool {
	return Tokens[second] > Tokens[first]
}

func SimplifyPotentialExpression(expression string, pass int, fileBase string, lineNumber int) string {
	expression = strings.TrimSpace(expression)

	if utils.StringInSlice(strings.ToUpper(expression), append(append(RegisterNames8, RegisterNames16...), ConditionCodes...)) {
		return expression
	}

	if expression[0] == '"' && expression[len(expression)-1] == '"' {
		// it's a string
		return expression
	}

	isIndirectAccess := false
	if expression[0] == '[' && expression[len(expression)-1] == ']' {
		// remove the brackets for now, add them back at the end
		isIndirectAccess = true
		expression = expression[1 : len(expression)-1]
	}

	tokens := ParseExpressionTokens(expression)

	// https://en.wikipedia.org/wiki/Shunting-yard_algorithm

	outputStack := []string{}
	operatorStack := []string{}
	poppedToken := ""

	for _, token := range tokens {
		num, ok := ParseNumber(token)
		definedVal, isDefinition := rom.Current.Definitions[token]
		if pass == 0 {
			// it's the first pass, so check if there's something that's in need of pointing
			if !isDefinition && utils.StringInSlice(token, rom.Current.UnpointedDefinitions) {
				// there is! use 0 as padding just so we can calculate where stuff is correctly
				// the actual value will be filled in on the second pass
				definedVal = 0
				isDefinition = true
			}
		}
		if ok {
			// the token is a number
			outputStack = append(outputStack, strconv.Itoa(num))
		} else if isDefinition {
			// the token is a number, but defined
			outputStack = append(outputStack, strconv.Itoa(definedVal))
		} else if token == "(" {
			operatorStack = append(operatorStack, token)
		} else if token == ")" {
			for operatorStack[len(operatorStack)-1] != "(" {
				poppedToken, operatorStack = operatorStack[len(operatorStack)-1], operatorStack[:len(operatorStack)-1]
				outputStack = append(outputStack, poppedToken)
			}
			if len(operatorStack) == 0 {
				log.Fatalf("Extra ')' at %s:%d", fileBase, lineNumber)
			}
			// remove the left parenthesis
			operatorStack = operatorStack[:len(operatorStack)-1]
		} else {
			// is there an operator with a greater precedence at the top?
			for len(operatorStack) > 0 && IsSecondOperatorMoreImportantThanFirst(token, operatorStack[len(operatorStack)-1]) {
				poppedToken, operatorStack = operatorStack[len(operatorStack)-1], operatorStack[:len(operatorStack)-1]
				outputStack = append(outputStack, poppedToken)
			}
			operatorStack = append(operatorStack, token)
		}
	}
	for len(operatorStack) > 0 {
		poppedToken, operatorStack = operatorStack[len(operatorStack)-1], operatorStack[:len(operatorStack)-1]
		if poppedToken == "(" || poppedToken == ")" {
			log.Fatalf("Extra '%s' at %s:%d", poppedToken, fileBase, lineNumber)
		}
		outputStack = append(outputStack, poppedToken)
	}

	rpnStack := []int{}
	for _, token := range outputStack {
		val, err := strconv.Atoi(token)
		if err == nil {
			// it's a number, add it to the stack
			rpnStack = append(rpnStack, val)
		} else {
			// it's an operand that requires 2 parameters
			if len(rpnStack) < 2 {
				log.Fatalf("Error parsing expression '%s' at %s:%d", expression, fileBase, lineNumber)
			}
			first, second := 0, 0
			second, rpnStack = rpnStack[len(rpnStack)-1], rpnStack[:len(rpnStack)-1]
			first, rpnStack = rpnStack[len(rpnStack)-1], rpnStack[:len(rpnStack)-1]
			if token == "+" {
				rpnStack = append(rpnStack, first+second)
			} else if token == "-" {
				rpnStack = append(rpnStack, first-second)
			} else if token == "*" {
				rpnStack = append(rpnStack, first*second)
			} else if token == "/" {
				rpnStack = append(rpnStack, first/second)
			} else if token == ">>" {
				rpnStack = append(rpnStack, int(uint(first)>>uint(second)))
			} else if token == "<<" {
				rpnStack = append(rpnStack, int(uint(first)<<uint(second)))
			} else if token == "|" {
				rpnStack = append(rpnStack, first|second)
			} else if token == "&" {
				rpnStack = append(rpnStack, first&second)
			} else if token == "^" {
				rpnStack = append(rpnStack, first&second)
			} else {
				log.Fatalf("Unknown token '%s' when parsing expression '%s' at %s:%d", token, expression, fileBase, lineNumber)
			}
		}
	}

	if len(rpnStack) == 0 || len(rpnStack) > 1 {
		log.Fatalf("Missing operands for expression '%s' at %s:%d", expression, fileBase, lineNumber)
	}

	result := strconv.Itoa(rpnStack[0])
	if isIndirectAccess {
		result = "[" + result + "]"
	}

	return result
}
