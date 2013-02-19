/*
    This file is part of Plotomaton.

    Plotomaton is free software: you can redistribute it and/or modify
    it under the terms of the GNU General Public License as published by
    the Free Software Foundation, either version 3 of the License, or
    (at your option) any later version.

    Plotomaton is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU General Public License for more details.

    You should have received a copy of the GNU General Public License
    along with Plotomaton.  If not, see <http://www.gnu.org/licenses/>.

	Now that that's out of the way:
	Original Author - Samuel Payson
	Since updated, and currently maintained, by Sean Anderson
	Contact: fnordit@gmail.com
*/

package main

import (
	"fmt"
	"parser"
	"math/rand"
	"os"
)

const debug bool = false

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("Please provide an input filename as an argument\n")
		return
	}

	input := os.Args[1]

	fmt.Printf("Running: %s\n", input)

	u := parser.ParseFile(input)
	if u == nil {
		fmt.Printf("Invalid input file\n")
		return
	}
	s := u.Instantiate()
	r := rand.New(rand.NewSource(rand.Int63()))

	log := s.Now()

	for {
		s.RunSpontaneous(r)

		// Display events
		for {
			log = log.Future()
			if log == nil {
				fmt.Printf("Error or something\n")
			}
			if cause := log.Cause(); cause != nil {
				fmt.Printf("%s\n", cause.Description())
			}
			if log == s.Now() {
				break
			}
		}

		if debug {
			p := s.PossibleTransitions()
			fmt.Printf("[%v can %v]\n", s, p)
		}

		// Have user choose a transition
		choices := s.ChosenTransitions()
		fmt.Fprintf(os.Stdout, "  0. Exit.\n")
		fmt.Fprintf(os.Stderr, "  1. Do nothing.\n")
		for i, t := range choices {
			fmt.Fprintf(os.Stderr, "  %d. %s\n", i+2, t.ChoiceDescription())
		}
		var choice int
		for {
			fmt.Fprintf(os.Stderr, "> ")
			fmt.Scanf("%d", &choice)
//			n, err := fmt.Scanf("%d", &choice)
/*			if n < 1 {
				if err.String() == "unexpected newline" || err.String() == "expected integer" {
					// TODO: KLUDGE. What we want here is distinguishing "malformed input" from "I/O error."
					// read again
				} else {
					fmt.Fprintf(os.Stderr, "error reading stdin: %s\n", err.String())
					return
				}
			} else*/ if choice < 0 || choice > len(choices) {
				fmt.Printf("Please enter a number between 1 and %d.\n", len(choices))
			} else /* good result */ {
				break
			}
		}
		if choice == 0 {
			return
		}
		if choice != 1 {
			choices[choice-1].Apply(s)
		}
	}
}
