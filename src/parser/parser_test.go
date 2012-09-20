package parser_test

import (
	"parser"
//    "state"
    "testing"
)

func assert(t *testing.T, name string, want interface{}, got interface{}) bool {
    r := want == got
    if !r {
        t.Error(name, " expected:", want, " got:", got)
    }
    return r
}

func Test_Parser(t *testing.T) {
	parser.ParseFile("test")
}
