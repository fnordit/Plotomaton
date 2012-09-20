package state_test

import (
	"state"
	"testing"
	"math/rand"
)

func assert(t *testing.T, name string, want interface{}, got interface{}) bool {
	r := want == got
	if !r {
		t.Error(name, " expected:", want, "got:", got)
	}
	return r
}

func initial() (*state.Universe, *rand.Rand, *state.Factor) {
	u := state.NewUniverse()
	r := rand.New(rand.NewSource(0))
	f := u.AddFactor("a-factor", "a", []string{"a", "b", "c"})

	return u, r, f
}

////////////////////////////////////////////////////////////////////////////////

func Test_NewUniverse(t *testing.T) {
	state.NewUniverse()
}

func Test_BoolExpr(t *testing.T) {
	u, _, f := initial()
	s := u.Instantiate()

	// FactorEquals
	tru := state.FactorEquals{f, "a"}
	fals := state.FactorEquals{f, "b"}
	assert(t, "FactorEquals 1", true, tru.Evaluate(s))
	assert(t, "FactorEquals 2", false, fals.Evaluate(s))

	// And
	assert(t, "And 1", true, state.MkAnd(tru, tru).Evaluate(s))
	assert(t, "And 2", false, state.MkAnd(tru, fals).Evaluate(s))
	assert(t, "And 3", true, state.MkAnd().Evaluate(s))

	// Or
	assert(t, "Or 1", false, state.MkOr(fals, fals).Evaluate(s))
	assert(t, "Or 2", true, state.MkOr(tru, fals).Evaluate(s))
	assert(t, "Or 3", false, state.MkOr().Evaluate(s))
}

func Test_Spontaneous(t *testing.T) {
	u, r, f := initial()
	u.AddTransition("a-transition",
		state.FactorEquals{f, "a"},
		state.Spontaneous{1},
		"",
		map[*state.Factor]state.Value{f: "b"})
	s := u.Instantiate()

	assert(t, "Initial state", state.Value("a"), s.Get(f))

	s.RunSpontaneous(r)

	assert(t, "Spontaneous happened", state.Value("b"), s.Get(f))
}

func Test_SpontaneousNot(t *testing.T) {
	u, r, f := initial()
	u.AddTransition("a-transition",
		state.FactorEquals{f, "a"},
		state.Spontaneous{0},
		"",
		map[*state.Factor]state.Value{f: "b"})
	s := u.Instantiate()

	assert(t, "Initial state", state.Value("a"), s.Get(f))

	s.RunSpontaneous(r)

	assert(t, "Spontaneous didn't happen", state.Value("a"), s.Get(f))
}

func Test_EventHistory(t *testing.T) {
	u, _, f := initial()

	tr1 := u.AddTransition("transition1",
		state.FactorEquals{f, "a"},
		state.Spontaneous{1},
		"AB happened.",
		map[*state.Factor]state.Value{f: "b"})
	tr2 := u.AddTransition("transition2",
		state.FactorEquals{f, "b"},
		state.Spontaneous{1},
		"BC happened.",
		map[*state.Factor]state.Value{f: "c"})
	s := u.Instantiate()

	assert(t, "Empty history", true, nil == s.Now().Past())

	h0 := s.Now()

	if assert(t, "Initial history exists", true, h0 != nil) {
		assert(t, "Initial history past terminated", true, h0.Past() == nil)
		assert(t, "Initial history future terminated", true, h0.Future() == nil)

		tr1.Apply(s)

		h1 := h0.Future()
		if assert(t, "History added", true, h1 != nil) {
			assert(t, "History linked pastward", h0, h1.Past())
			assert(t, "History terminated", true, h1.Future() == nil)

			tr2.Apply(s)

			h2 := h1.Future()

			assert(t, "Initial state cause null", true, nil == h0.Cause())
			assert(t, "Transition 1 description", "AB happened.", h1.Cause().Description())
			assert(t, "Transition 2 description", "BC happened.", h2.Cause().Description())
		}
	}
}

func Test_UndoRedo(t *testing.T) {
	u, _, f := initial()
	tr1 := u.AddTransition("transition1",
		state.FactorEquals{f, "a"},
		state.Spontaneous{0},
		"AB happened.",
		map[*state.Factor]state.Value{f: "b"})
	tr2 := u.AddTransition("transition2",
		state.FactorEquals{f, "b"},
		state.Spontaneous{0},
		"BC happened.",
		map[*state.Factor]state.Value{f: "c"})
	s := u.Instantiate()

	assert(t, "No undo before do", false, nil != s.Now().Past())
	assert(t, "No redo before do", false, nil != s.Now().Future())
	tr1.Apply(s)
	tr2.Apply(s)
	assert(t, "Can undo after do", true, nil != s.Now().Past())
	assert(t, "Pre-undo", state.Value("c"), s.Get(f))
	assert(t, "Redo before mid", false, nil != s.Now().Future())
	s.Goto(s.Now().Past())
	assert(t, "Undo 1", state.Value("b"), s.Get(f))
	assert(t, "Redo after mid", true, nil != s.Now().Future())
	s.Goto(s.Now().Past())
	assert(t, "Undo 2", state.Value("a"), s.Get(f))
	s.Goto(s.Now().Future())
	assert(t, "Redo", state.Value("b"), s.Get(f))
}

// TODO: test AddFactor once there's a good way to inspect it
// TODO: test AddTransition once there's a good way to inspect it
// TODO: test Instantiate initial contents once there's a good way to inspect it
// TODO: test possible-transition calculation
