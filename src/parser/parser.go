package parser

import (
	"os"
	"bufio"
	"bytes"
	"state"
	"strconv"
)

const (
	EOF            = 255
	INT            = 128
	STRING         = 129
	FACTOR         = 130
	TRANSITION     = 131
	DESCRIPTION    = 132
	SPONTANEOUS    = 133
	CHOICE         = 134
	STRING_LITERAL = 135
	FLOAT          = 136
)

var file_reader *(bufio.Reader)
var current_token byte
var current_string string
var current_int int
var current_float float64

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
		case ':', '(', ')', ',', '<', '>', '=', '-', '\\', '+', '|', '&':
			return current_byte
        case '%':
            for (current_byte != '\n') {
                current_byte, err = file_reader.ReadByte()
            }
            return GetNextToken()
		case '"':
			current_buffer := bytes.NewBuffer(make([]byte, 0, 80))
			current_byte, _ := file_reader.ReadByte()
			for current_byte != '"' {
				current_buffer.WriteByte(current_byte)
				current_byte, _ = file_reader.ReadByte()
			}
			current_string = current_buffer.String()
			return STRING_LITERAL
		}
		if IsNum(current_byte) || current_byte == '.' {
			// current_byte is a digit - react accordingly
			current_int = 0
			for IsNum(current_byte) {
				current_int *= 10
				current_int += (int(current_byte) - 48)
				current_byte, err = file_reader.ReadByte()
			}
			current_float = float64(current_int)

			if current_byte == '.' {
				dec_place := 0.1
				current_byte, err = file_reader.ReadByte()
				for IsNum(current_byte) {
					current_float += float64(int(current_byte) - 48)*dec_place
					dec_place /= 10
					current_byte, err = file_reader.ReadByte()
				}
				file_reader.UnreadByte()
				return FLOAT
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
			default:
				return STRING
			}
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
	//var initial string
	name := FactorName()
	Match(':')
	Match('(')
	values := FactorValues()
	u.AddFactor(name, values[0], values)
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

func Transition() {
	Match(TRANSITION)
	var name string
	if current_token == STRING {
		name = TransitionName()
	} else {
		name = ""
	}
	Match(':')
	Match('(')
	expression := Conjunction()
	Match(',')
	schedule := Schedule()
	Match(',')
	effects := FactorTransitions()
	var description string
	if current_token == ',' {
		Match(',')
		description = current_string
		Match(STRING_LITERAL)
	} else {
		description = ""
	}
	Match(')')
	u.AddTransition(name, expression, schedule, description, effects)
}

func Schedule() state.Schedule {
	var ret state.Schedule
	if current_token == SPONTANEOUS {
		Match(SPONTANEOUS)
		if current_token == INT {
			ret = state.Spontaneous{float64(current_int)}
			Match(INT)
		} else {
			ret = state.Spontaneous{current_float}
			Match(FLOAT)
		}
	} else {
		Match(CHOICE)
		Match(':')
		ret = state.Chosen{current_string}
		Match(STRING_LITERAL)
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
func Conjunction() state.BoolExpr {
	var exps []state.BoolExpr
	exps = append(exps, Disjunction())
	for current_token == '&' {
		Match('&')
		exps = append(exps, Disjunction())
	}
	return state.And{exps}
}

func Disjunction() state.BoolExpr {
	var exps []state.BoolExpr
	exps = append(exps, Bool())
	for current_token == '|' {
		Match('|')
		exps = append(exps, Bool())
	}
	return state.Or{exps}
}

func Bool() state.BoolExpr {
	if current_token == '(' {
		Match('(')
		exp := Conjunction()
		Match(')')
		return exp
	} else {
		fac := u.FindFactor(FactorName())
		switch current_token {
//		case '<':
//			Match('<')
//			if current_token == '=' {
//				Match('=')
//			}

//			if current_token == INT {
//				Match(INT)
//			} else {
//				FactorName()
//			}
//		case '>':
//			Match('>')
//			if current_token == '=' {
//				Match('=')
//			}

//			if current_token == INT {
//				Match(INT)
//			} else {
//				FactorName()
//			}
		case '=':
			Match('=')
			if current_token == INT {
				exp := state.FactorEquals{fac, state.Value(strconv.Itoa(current_int))}
				Match(INT)
				return exp
			} else {
				exp := state.FactorEquals{fac, state.Value(current_string)}
				Match(STRING)
				return exp
			}
		}
	}
	return nil
}

func FactorTransitions() map[*state.Factor]state.Value {
	ret := make(map[*state.Factor]state.Value)
	if current_token != '(' {
		fac, val := FactorTransition()
		ret[fac] = val
	} else {
		Match('(')
		transitions := FactorTransitionList()
		for fac, val := range(transitions) {
			ret[fac] = val
		}
		Match(')')
	}
	return ret
}

func FactorTransitionList() map[*state.Factor]state.Value {
	ret := make(map[*state.Factor]state.Value)
	fac, val := FactorTransition()
	ret[fac] = val
	for current_token == ',' {
		Match(',')
		fac, val = FactorTransition()
		ret[fac] = val
	}
	return ret
}

func FactorTransition() (*state.Factor, state.Value) {
	fac := u.FindFactor(FactorName())
	var val state.Value
	if current_token == '-' {
		Match('-')
		if current_token == '>' {
			Match('>')
			val = state.Value(FactorValue())
		}// else {
//			Match(INT)
//		}
//	} else {
//		Match('+')
//		if current_token == INT {
//			Match(INT)
//		}
	}
	return fac, val
}


func Description() {
	Match(':')
	Match('(')
	Conjunction()
	Match(',')
//	StringLiteral()
	Match(STRING_LITERAL)
	Match(')')
}
