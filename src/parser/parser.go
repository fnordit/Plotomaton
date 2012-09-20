package parser

import (
	"os"
	"bufio"
	"bytes"
	"state"
    "fmt"
)

const (
	EOF         = 255
	INT         = 128
	STRING      = 129
	FACTOR      = 130
	TRANSITION  = 131
	DESCRIPTION = 132
	SPONTANEOUS = 133
	CHOICE      = 134
)

var file_reader *(bufio.Reader)
var current_token byte
var current_string string
var current_int int

func error() {

}

func Match(b byte) {
	if b != current_token {
		error()
	} else {
		current_token = GetNextToken()
	}
	return
}

func IsAlpha(c byte) bool {
	return ((c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || c == '_')
}

func IsNum(c byte) bool {
	return (c >= '0' && c <= '9')
}

func GetNextToken() byte {
	current_byte, err := file_reader.ReadByte()
	if err != nil {
		return EOF
	} else {
		switch current_byte {
		case ' ', '\n':
			return GetNextToken()
		case ':', '(', ')', ',', '<', '>', '=', '-', '\\', '"', '+', '|', '&':
			return current_byte
        case '%':
            for (current_byte != '\n') {
                current_byte, err = file_reader.ReadByte()
            }
            return GetNextToken()
		}
		if IsNum(current_byte) {
			// current_byte is a digit - react accordingly
			current_int = 0
			for IsNum(current_byte) {
				current_int *= 10
				current_int += (int(current_byte) - 48)
				current_byte, err = file_reader.ReadByte()
			}
			file_reader.UnreadByte()
			return INT
		} else if IsAlpha(current_byte) {
			// current_byte is a letter
			current_buffer := bytes.NewBuffer(make([]byte, 0, 80))
			for IsAlpha(current_byte) || IsNum(current_byte) {
				current_buffer.WriteByte(current_byte)
				current_byte, err = file_reader.ReadByte()
			}
			file_reader.UnreadByte()
			current_string = current_buffer.String()
			switch {
			case current_string == "factor":
				return FACTOR
			case current_string == "transition":
				return TRANSITION
			case current_string == "description":
				return DESCRIPTION
			case current_string == "spontaneous":
				return SPONTANEOUS
			case current_string == "choice":
				return CHOICE
			}
			return STRING
		}
	}
	return 0
}

var u *state.Universe

func ParseFile(filename string) *state.Universe {
	f, err := os.Open(filename)
	file_reader = bufio.NewReader(f)
	if err != nil {
		return nil
	}

	defer f.Close()

	u = state.NewUniverse()

	current_token = GetNextToken()
	AllFile()

	return u
}

func AllFile() {
	for current_token != EOF {
		switch current_token {
		case FACTOR:
			Match(FACTOR)
			Factor()
		case TRANSITION:
			Match(TRANSITION)
			Transition()
		case DESCRIPTION:
			Match(DESCRIPTION)
			Description()
		}
	}

	return
}

func Factor() {
	var initial string
	name := FactorName()
	Match(':')
	Match('(')
	values := FactorValues()
	u.AddFactor(name, initial, values)
	Match(')')
}

func FactorName() string {
	var name string
	if current_token == STRING {
		name = current_string
		Match(STRING)
	} else {
		error()
	}
	return name
}

func FactorValues() []string {
	vals := make([]string, 1)
	if current_token == STRING {
		vals[0] = FactorValue()
		if current_token == ',' {
			Match(',')
			vals = append(vals, FactorValues()...)
		}
	} /* else if (current_token == INT) {
	       first := current_int
	       Match(INT)
	       Match(':')
	       second := current_int
	       Match(INT)
	   }
	*/
	return vals
}

func FactorValue() string {
	val := current_string
	Match(STRING)
	return val
}

func Transition() state.Transition {
	Match(TRANSITION)
	var name string
	if current_token == STRING {
		name = TransitionName()
	} else {
		name = ""
	}
    fmt.Printf("%s\n", name)
	Match(':')
	Match('(')
	Conjunction()
	Match(',')
	Schedule()
	Match(',')
	FactorTransitions()
	var description string
	if current_token == ',' {
		Match(',')
		description = StringLiteral()
        fmt.Printf("%s\n", description)
	} else {
		description = ""
	}
	Match(')')
    tran := state.Transition{name, nil, nil, description, nil}
}

func Schedule() state.Schedule {
	var ret state.Schedule
	if current_token == SPONTANEOUS {
		Match(SPONTANEOUS)
		ret = state.Spontaneous{float64(current_int)}
		Match(INT)
	} else {
		Match(CHOICE)
		ret = state.Chosen{StringLiteral()}
	}
	return ret
}

func TransitionName() string {
	name := current_string
	Match(STRING)
	return name
}

// Boolean requirements for the execution of transitions
// Written in conjunctive form - CLAUSE & CLAUSE & ...
func Conjunction() {
	Disjunction()
	for current_token == '&' {
		Match('&')
		Disjunction()
	}
}

func Disjunction() {
	Bool()
	for current_token == '|' {
		Match('|')
		Bool()
	}
}

func Bool() {
	if current_token == '(' {
		Match('(')
		Conjunction()
		Match(')')
	} else {
		FactorName()
		switch current_token {
		case '<':
			Match('<')
			if current_token == '=' {
				Match('=')
			}

			if current_token == INT {
				Match(INT)
			} else {
				FactorName()
			}
		case '>':
			Match('>')
			if current_token == '=' {
				Match('=')
			}

			if current_token == INT {
				Match(INT)
			} else {
				FactorName()
			}
		case '=':
			Match('=')
			if current_token == INT {
				Match(INT)
			} else {
				Match(STRING)
			}
		}
	}
}

func FactorTransitions() {
	if current_token != '(' {
		FactorTransition()
	} else {
		Match('(')
		FactorTransitionList()
		Match(')')
	}
}

func FactorTransitionList() {
	FactorTransition()
	for current_token == ',' {
		Match(',')
		FactorTransition()
	}
}

func FactorTransition() {
	FactorName()
	if current_token == '-' {
		Match('-')
		if current_token == '>' {
			Match('>')
			FactorValue()
		} else {
			Match(INT)
		}
	} else {
		Match('+')
		if current_token == INT {
			Match(INT)
		}
	}
}

func StringLiteral() string {
	if current_token != '"' {
		error()
	}
	var previous_char byte
	var current_char byte
	current_char = ' '
	ret := ""
	for current_char != '"' || previous_char == '\\' {
		previous_char = current_char
		current_char, _ = file_reader.ReadByte()
		ret += string(current_char)
	}
	file_reader.UnreadByte()
	current_token = '"'
	Match('"')
	Match('"')
	return ret
}

func Description() {
	Match(':')
	Match('(')
	Conjunction()
	Match(',')
	StringLiteral()
	Match(')')
}
