package main

import (
	"fmt"
	"state"
	"math/rand"
	"os"
)

const debug bool = false

func main() {
	u := state.NewSampleUniverse()
	s := u.Instantiate()
	r := rand.New(rand.NewSource(rand.Int63()))

	log := s.Now()

	if debug {
		fmt.Printf("Hello, %v.\n", u)
	}

	for {
		s.RunSpontaneous(r)

		// Display events
		for {
			log = log.Future()
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
		fmt.Fprintf(os.Stderr, "  0. Do nothing.\n")
		for i, t := range choices {
			fmt.Fprintf(os.Stderr, "  %d. %s\n", i+1, t.ChoiceDescription())
		}
		var choice int
		for {
			fmt.Fprintf(os.Stderr, "> ")
			n, err := fmt.Scanf("%d", &choice)
			if n < 1 {
				if err.String() == "unexpected newline" || err.String() == "expected integer" {
					// TODO: KLUDGE. What we want here is distinguishing "malformed input" from "I/O error."
					// read again
				} else {
					fmt.Fprintf(os.Stderr, "error reading stdin: %s\n", err.String())
					return
				}
			} else if choice < 0 || choice > len(choices) {
				fmt.Printf("Please enter a number between 1 and %d.\n", len(choices))
			} else /* good result */ {
				break
			}
		}
		if choice != 0 {
			choices[choice-1].Apply(s)
		}
	}
}
