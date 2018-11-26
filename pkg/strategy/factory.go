package strategy

import (
	"fmt"
	"log"
	"sync"

	"github.com/curlymon/kicker/pkg/conf"
	"k8s.io/api/core/v1"
)

// Strategy is an object used to statefully evaluate a set of v1.Pod for to be kicked
type Strategy struct {
	c    conf.Criteria
	eval func([]v1.Pod) []v1.Pod
}

// Criteria return the conf.Criteria used to create this Strategy
func (s *Strategy) Criteria() conf.Criteria {
	return s.c
}

// Evaluate performs the evaluation defined by the conf.Criteria used to create this Strategy
func (s *Strategy) Evaluate(pods []v1.Pod) []v1.Pod {
	log.Printf("Evaluate called with %d pods", len(pods))
	return s.eval(pods)
}

// NewStrategy builds and returns a new Strategy for the provided conf.Criteria, returning an error if unable to do so.
func NewStrategy(c conf.Criteria) (*Strategy, error) {
	stratCon, err := RetrieveEvaluatorConstructor(c.Strategy)
	if err != nil {
		return nil, err
	}

	return &Strategy{
		c:    c,
		eval: stratCon(c),
	}, nil
}

// NewGroup is a convenience function to create a group of strategies in a single call
func NewGroup(cs []conf.Criteria) ([]*Strategy, error) {
	strats := make([]*Strategy, 0, len(cs))
	for i := range cs {
		strat, err := NewStrategy(cs[i])
		if err != nil {
			return nil, err
		}

		strats = append(strats, strat)
	}

	return strats, nil
}

// Evaluator is the logic kernel used to evaluate kicking a set of pods
type Evaluator func([]v1.Pod) []v1.Pod

// EvaluatorConstructor defines a constructor function for an Evaluator
type EvaluatorConstructor func(conf.Criteria) Evaluator

var strategyRegistry = map[conf.Strategy]EvaluatorConstructor{}
var mu = &sync.RWMutex{}

// RegisterEvaluatorConstructor registers an EvaluatorConstructor for use with a given name. This can then be referenced
// from a conf.Criteria.Strategy for use.
func RegisterEvaluatorConstructor(strat conf.Strategy, con EvaluatorConstructor) error {
	mu.Lock()
	defer mu.Unlock()
	if _, ok := strategyRegistry[strat]; ok {
		return fmt.Errorf("strategy '%s' is already registered", strat)
	}

	strategyRegistry[strat] = con

	return nil
}

// RetrieveEvaluatorConstructor retrieves a registered EvaluatorConstructor.
func RetrieveEvaluatorConstructor(strat conf.Strategy) (EvaluatorConstructor, error) {
	mu.RLock()
	defer mu.RUnlock()
	stratCon, ok := strategyRegistry[strat]
	if !ok {
		return nil, fmt.Errorf("strategy '%s' is not registered for use", strat)
	}

	return stratCon, nil
}
