package state

import (
	"strings"
	"math/rand"
)

////////////////////////////////////////////////////////////////////////////////

// A Universe is a collection of Factors. Each Factor has a set of possible
// Values and an initial Value. A State of a Universe gives “current” Values
// for every Factor.

type Value string

type Universe struct {
	factors     map[*Factor]bool
	transitions map[*Transition]bool
}

type Factor struct {
	label    string
	initial  Value
	possible map[Value]bool
}

type State struct {
	universe *Universe
	now      *Moment
}

// A Transition defines a change in the State. Each transition has:
//   a label, for the author's use
//   a condition, which determines what states it can occur in
//   a schedule, which determines when it occurs when it can
//   a description, which the player sees when it happens
//   effects, which are modifications to the state
type Transition struct {
	label       string
	condition   BoolExpr
	schedule    Schedule
	description string
	effects     map[*Factor]Value
}

type Schedule interface {
	now(*rand.Rand) bool
	ask() bool
	ChoiceDescription() string
}

type Spontaneous struct {
	ProbabilityPerTurn float64
}

type Chosen struct {
	Description string
}

type Moment struct {
	universe *Universe
	values   map[*Factor]Value
	future   *Moment // TODO should be read-only
	past     *Moment
	cause    *Transition
}

////////////////////////////////////////////////////////////////////////////////

func NewUniverse() *Universe {
	return &Universe{map[*Factor]bool{}, map[*Transition]bool{}}
}

func (u Universe) String() string {
	return listing("Universe[", "]", func(write func(string)) {
		for f, _ := range u.factors {
			write(f.String())
		}
	})
}

// Note: Result is in an invalid state as its initial value is not a possible
// value (and it has no possible values).
func newFactor(label string) *Factor {
	return &Factor{label, Value(""), map[Value]bool{}}
}

func (f Factor) String() string {
	return listing("{"+f.label+" [", "] "+string(f.initial)+"}", func(write func(string)) {
		for v, _ := range f.possible {
			write(string(v))
		}
	})
}

func (u *Universe) AddFactor(label string, initial string, values []string) *Factor {
	f := newFactor(label)
	for _, v := range values {
		f.possible[Value(v)] = true
	}
	f.initial = Value(initial)
	u.factors[f] = true
	return f
}

func (u *Universe) AddTransition(label string, condition BoolExpr, schedule Schedule, description string, effects map[*Factor]Value) *Transition {
	// TODO: deepcopy maps or otherwise avoid aliasing
	t := &Transition{label, condition, schedule, description, effects}
	u.transitions[t] = true
	return t
}

func newState(u *Universe, initial map[*Factor]Value) *State {
	var s State
	var m Moment
	s.universe = u
	s.now = &m
	m.universe = u
	m.values = copyMap(initial)
	return &s
}

// Create a State of this Universe with initial values.
func (u *Universe) Instantiate() *State {
	vs := map[*Factor]Value{}
	for f, _ := range u.factors {
		vs[f] = f.initial
	}
	return newState(u, vs)
}

func (s State) String() string {
	u := s.universe
	return listing("State[", "]", func(write func(string)) {
		for f, _ := range u.factors {
			v := s.now.values[f]
			_, valid := f.possible[v]
			var note string
			if valid {
				note = ""
			} else {
				note = "<INVALID>"
			}
			write(f.label + ":" + note + string(v))
		}
	})
}

// TODO: Should this return something finer than just a Transition?
func (s *State) PossibleTransitions() []*Transition {
	var ts []*Transition
	for t, _ := range s.universe.transitions {
		if t.condition.Evaluate(s) {
			ts = append(ts, t)
		}
	}
	return ts
}

func (s *State) RunSpontaneous(r *rand.Rand) {
	// TODO: Need to do this loop in random order for unbiased operation.
	for _, t := range s.PossibleTransitions() {
		// The condition is reevaluated in case a previously run transition changed things.
		if t.schedule.now(r) && t.condition.Evaluate(s) {
			//fmt.Printf("[Running spontaneous transition %v]\n", t)
			t.Apply(s)
		}
	}
}

// Return all user-selectable transitions for the current state.
func (s *State) ChosenTransitions() []*Transition {
	var ts []*Transition
	for _, t := range s.PossibleTransitions() {
		if t.schedule.ask() {
			ts = append(ts, t)
		}
	}
	return ts
}

func (s *State) Get(f *Factor) Value {
	return s.now.values[f]
}

func (t Transition) String() string {
	return "{" + t.label + "...}"
}

func (t Transition) Description() string {
	return t.description
}

func (t Transition) ChoiceDescription() string {
	return t.schedule.ChoiceDescription()
}

func (t *Transition) Apply(s *State) {
	var newNow Moment
	newNow.universe = s.universe
	newNow.cause = t

	newNow.values = copyMap(s.now.values)
	for f, v := range t.effects {
		newNow.values[f] = v
	}

	newNow.past = s.now
	s.now.future = &newNow

	s.now = &newNow
}

// interface methods
func (s Spontaneous) now(r *rand.Rand) bool {
	return r.Float64() <= s.ProbabilityPerTurn
}
func (_ Spontaneous) ask() bool                 { return false }
func (_ Spontaneous) ChoiceDescription() string { return "Missingno" }
func (_ Chosen) now(r *rand.Rand) bool          { return false }
func (_ Chosen) ask() bool                      { return true }
func (c Chosen) ChoiceDescription() string      { return c.Description }

////////////////////////////////////////////////////////////////////////////////

// Boolean expressions, used for transition conditions

type BoolExpr interface {
	Evaluate(s *State) bool
}

type FactorEquals struct {
	Factor *Factor
	Value  Value
}

type And struct {
	Clauses []BoolExpr
}

type Or struct {
	Clauses []BoolExpr
}

func MkAnd(clauses ...BoolExpr) BoolExpr {
	return And{clauses}
}

func MkOr(clauses ...BoolExpr) BoolExpr {
	return Or{clauses}
}

func (e FactorEquals) Evaluate(s *State) bool {
	return s.now.values[e.Factor] == e.Value
}

func (e And) Evaluate(s *State) bool {
	for _, e := range e.Clauses {
		if !e.Evaluate(s) {
			return false
		}
	}
	return true
}

func (e Or) Evaluate(s *State) bool {
	for _, e := range e.Clauses {
		if e.Evaluate(s) {
			return true
		}
	}
	return false
}

////////////////////////////////////////////////////////////////////////////////

// History access

func (s *State) Now() *Moment {
	return s.now
}

func (s *State) Goto(m *Moment) {
	s.now = m
}

func (m Moment) Future() *Moment {
	return m.future
}

func (m Moment) Past() *Moment {
	return m.past
}

func (m Moment) Cause() *Transition {
	return m.cause
}

////////////////////////////////////////////////////////////////////////////////

func copyMap(in map[*Factor]Value) map[*Factor]Value {
	out := make(map[*Factor]Value)
	for k, v := range in {
		out[k] = v
	}
	return out
}

// utility for printing lists
func listing(prefix string, suffix string, body func(func(string))) string {
	var s []string
	s = append(s, prefix)
	first := true
	body(func(v string) {
		if !first {
			s = append(s, ", ")
		} else {
			first = false
		}
		s = append(s, v)
	})
	s = append(s, suffix)
	return strings.Join(s, "")
}

////////////////////////////////////////////////////////////////////////////////

// Example universe; a placeholder until we have a parser for game data files

func NewSampleUniverse() *Universe {
	u := NewUniverse()

	key := u.AddFactor("LabKey", "no", []string{"yes", "no"})

	sun := u.AddFactor("sun", "day", []string{"day", "night"})
	u.AddTransition("sunrise",
		FactorEquals{sun, "night"},
		Spontaneous{1},
		"The sun rises.",
		map[*Factor]Value{sun: "day"})
	u.AddTransition("sunset",
		FactorEquals{sun, "day"},
		Spontaneous{1},
		"The sun sets.",
		map[*Factor]Value{sun: "night"})

	loc := u.AddFactor("location", "COSI", []string{"Hallway", "COSI", "ITL"})
	u.AddTransition("ToCOSI",
		MkOr(
			MkAnd(
				FactorEquals{loc, "Hallway"},
				FactorEquals{key, "yes"}),
			FactorEquals{loc, "ITL"}),
		Chosen{"Enter COSI."},
		"You walk into the COSI lab.",
		map[*Factor]Value{loc: "COSI"})
	u.AddTransition("ToITL",
		MkOr(
			MkAnd(
				FactorEquals{loc, "Hallway"},
				FactorEquals{key, "yes"}),
			FactorEquals{loc, "COSI"}),
		Chosen{"Enter ITL."},
		"You walk into the ITL.",
		map[*Factor]Value{loc: "ITL"})
	u.AddTransition("ToHallway",
		MkOr(
			FactorEquals{loc, "ITL"},
			FactorEquals{loc, "COSI"}),
		Chosen{"Leave room."},
		"You step out into the hallway.",
		map[*Factor]Value{loc: "Hallway"})

	return u
}
