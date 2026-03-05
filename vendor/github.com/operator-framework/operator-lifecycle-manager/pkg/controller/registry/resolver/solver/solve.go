package solver

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/go-air/gini"
	"github.com/go-air/gini/inter"
	"github.com/go-air/gini/z"
)

var ErrIncomplete = errors.New("cancelled before a solution could be found")

// NotSatisfiable is an error composed of a minimal set of applied
// constraints that is sufficient to make a solution impossible.
type NotSatisfiable []AppliedConstraint

func (e NotSatisfiable) Error() string {
	const msg = "constraints not satisfiable"
	if len(e) == 0 {
		return msg
	}
	s := make([]string, len(e))
	for i, a := range e {
		s[i] = a.String()
	}
	return fmt.Sprintf("%s: %s", msg, strings.Join(s, ", "))
}

type Solver interface {
	Solve(context.Context) ([]Variable, error)
}

type solver struct {
	g      inter.S
	litMap *litMapping
	tracer Tracer
	buffer []z.Lit
	log    io.Writer // optional writer for RESOLUTION_TRACE lines
}

const (
	satisfiable   = 1
	unsatisfiable = -1
	unknown       = 0
)

// Solve takes a slice containing all Variables and returns a slice
// containing only those Variables that were selected for
// installation. If no solution is possible, or if the provided
// Context times out or is cancelled, an error is returned.
func (s *solver) solverTrace(phase, event string, kvs ...string) {
	if s == nil || s.log == nil {
		return
	}
	fmt.Fprintf(s.log, "RESOLUTION_TRACE phase=%s event=%s %s\n", phase, event, strings.Join(kvs, " "))
}

func (s *solver) solverTraceBegin(phase string, kvs ...string) time.Time {
	s.solverTrace(phase, "BEGIN", kvs...)
	return time.Now()
}

func (s *solver) solverTraceDone(phase string, start time.Time, kvs ...string) {
	ms := time.Since(start).Milliseconds()
	s.solverTrace(phase, "DONE", append([]string{fmt.Sprintf("duration_ms=%d", ms)}, kvs...)...)
}

func (s *solver) Solve(ctx context.Context) (result []Variable, err error) {
	defer func() {
		// This likely indicates a bug, so discard whatever
		// return values were produced.
		if derr := s.litMap.Error(); derr != nil {
			result = nil
			err = derr
		}
	}()

	// teach all constraints to the solver
	variableCount := len(s.litMap.inorder)
	constraintCount := len(s.litMap.constraints)
	literalCount := len(s.litMap.lits)
	addStart := s.solverTraceBegin("sat_add_constraints",
		fmt.Sprintf("variable_count=%d", variableCount),
		fmt.Sprintf("constraint_count=%d", constraintCount),
		fmt.Sprintf("literal_count=%d", literalCount))
	s.litMap.AddConstraints(s.g)
	s.solverTraceDone("sat_add_constraints", addStart,
		fmt.Sprintf("variable_count=%d", variableCount),
		fmt.Sprintf("constraint_count=%d", constraintCount),
		fmt.Sprintf("literal_count=%d", literalCount))

	// collect literals of all mandatory variables to assume as a baseline
	var assumptions []z.Lit
	for _, anchor := range s.litMap.AnchorIdentifiers() {
		assumptions = append(assumptions, s.litMap.LitOf(anchor))
	}
	anchorCount := len(assumptions)

	// assume that all constraints hold
	s.litMap.AssumeConstraints(s.g)
	s.g.Assume(assumptions...)

	var aset map[z.Lit]struct{}
	// push a new test scope with the baseline assumptions, to prevent them from being cleared during search
	outcome, _ := s.g.Test(nil)
	if outcome != satisfiable && outcome != unsatisfiable {
		// searcher for solutions in input order, so that preferences
		// can be taken into acount (i.e. prefer one catalog to another)
		searchStart := s.solverTraceBegin("sat_search",
			fmt.Sprintf("anchor_count=%d", anchorCount),
			fmt.Sprintf("assumption_count=%d", len(assumptions)))
		outcome, assumptions, aset = (&search{s: s.g, lits: s.litMap, tracer: s.tracer}).Do(context.Background(), assumptions)
		s.solverTraceDone("sat_search", searchStart,
			fmt.Sprintf("outcome=%d", outcome),
			fmt.Sprintf("anchor_count=%d", anchorCount),
			fmt.Sprintf("final_assumption_count=%d", len(assumptions)))
	}
	switch outcome {
	case satisfiable:
		minimizeStart := s.solverTraceBegin("sat_minimize")
		s.buffer = s.litMap.Lits(s.buffer)
		var extras, excluded []z.Lit
		for _, m := range s.buffer {
			if _, ok := aset[m]; ok {
				continue
			}
			if !s.g.Value(m) {
				excluded = append(excluded, m.Not())
				continue
			}
			extras = append(extras, m)
		}
		s.g.Untest()
		cs := s.litMap.CardinalityConstrainer(s.g, extras)
		s.g.Assume(assumptions...)
		s.g.Assume(excluded...)
		s.litMap.AssumeConstraints(s.g)
		_, s.buffer = s.g.Test(s.buffer)
		for w := 0; w <= cs.N(); w++ {
			s.g.Assume(cs.Leq(w))
			if s.g.Solve() == satisfiable {
				vars := s.litMap.Variables(s.g)
				s.solverTraceDone("sat_minimize", minimizeStart,
					fmt.Sprintf("extras_count=%d", len(extras)),
					fmt.Sprintf("excluded_count=%d", len(excluded)),
					fmt.Sprintf("cardinality_bound=%d", w),
					fmt.Sprintf("result_count=%d", len(vars)))
				return vars, nil
			}
		}
		s.solverTraceDone("sat_minimize", minimizeStart,
			fmt.Sprintf("extras_count=%d", len(extras)),
			fmt.Sprintf("excluded_count=%d", len(excluded)),
			"error=true")
		// Something is wrong if we can't find a model anymore
		// after optimizing for cardinality.
		return nil, fmt.Errorf("unexpected internal error")
	case unsatisfiable:
		return nil, NotSatisfiable(s.litMap.Conflicts(s.g))
	}

	return nil, ErrIncomplete
}

func New(options ...Option) (Solver, error) {
	s := solver{g: gini.New()}
	for _, option := range append(options, defaults...) {
		if err := option(&s); err != nil {
			return nil, err
		}
	}
	return &s, nil
}

type Option func(s *solver) error

func WithInput(input []Variable) Option {
	return func(s *solver) error {
		var err error
		s.litMap, err = newLitMapping(input)
		return err
	}
}

func WithTracer(t Tracer) Option {
	return func(s *solver) error {
		s.tracer = t
		return nil
	}
}

func WithLogWriter(w io.Writer) Option {
	return func(s *solver) error {
		s.log = w
		return nil
	}
}

var defaults = []Option{
	func(s *solver) error {
		if s.litMap == nil {
			var err error
			s.litMap, err = newLitMapping(nil)
			return err
		}
		return nil
	},
	func(s *solver) error {
		if s.tracer == nil {
			s.tracer = DefaultTracer{}
		}
		return nil
	},
}
