package immediate

import (
	"log"
	"time"

	"github.com/curlymon/kicker/pkg/conf"
	"github.com/curlymon/kicker/pkg/strategy"
	"k8s.io/api/core/v1"
)

func init() {
	if err := strategy.RegisterEvaluatorConstructor(conf.StrategyImmediate, Immediate); err != nil {
		log.Fatal(err)
	}
}

// Immediate defines the immediate strategy evaluation.
// It first filters the passed list of pods to the setting defined in the passed conf.Criteria.
// Then it sorts pods oldest to newest by v1.Pod.CreationTimestamp.Time.
// Last it iterativly looks at the list of pods in order and evaluates if it should be kicked; adding this to a list
// until conf.Criteria.Limit is reached.
// If a pod is kicked, a cooldown for re-evaluation is triggered with a length of conf.Criteria.CoolDown.
func Immediate(c conf.Criteria) strategy.Evaluator {
	// build a logic core that assumes filtered pods
	core := strategy.EvaluatorSeive(
		strategy.SortCreationTimestampAsc,
		strategy.OlderThan(time.Duration(c.MaxAge)*time.Second),
		strategy.Limit(c.Limit),
	)

	// build a filter top remove all non matching and unhealthy pods
	filter := strategy.And(
		strategy.NamePrefixFilter(c.Name),
		strategy.NameSpaceFilter(c.Namespace),
		strategy.StatusFilter(v1.PodRunning),
	)

	// setup prefilter
	prefilter := strategy.EvaluatorSeive(
		strategy.ApplyFilter(filter),
		core,
	)

	// wrap prefilter strategy with cooldown
	return strategy.CoolDown(time.Duration(c.CoolDown)*time.Second, prefilter)
}
