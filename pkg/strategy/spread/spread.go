package spread

import (
	"log"
	"time"

	"github.com/curlymon/kicker/pkg/conf"
	"github.com/curlymon/kicker/pkg/strategy"
	"k8s.io/api/core/v1"
)

func init() {
	if err := strategy.RegisterEvaluatorConstructor(conf.StrategySpread, Spread); err != nil {
		log.Fatal(err)
	}
}

// Spread defines the spread strategy evaluation.
// It first filters the passed list of pods to the setting defined in the passed conf.Criteria.
// Then it sorts pods oldest to newest by v1.Pod.CreationTimestamp.Time.
// Last it iterativly looks at the list of pods in order and evaluates if it should be kicked; adding this to a list
// until conf.Criteria.Limit is reached.
// If a pod is kicked, a cooldown for re-evaluation is triggered with a length of conf.Criteria.CoolDown to prevent
// scheduler thrash. An additional cooldown is triggered for (conf.Criteria.MaxAge / podCount), this cooldown is ignored
// if the pod count is higher then the last count at pod kick
func Spread(c conf.Criteria) strategy.Evaluator {
	// fiter out non matching and unhealthy pods
	filter := strategy.And(
		strategy.NamePrefixFilter(c.Name),
		strategy.NameSpaceFilter(c.Namespace),
		strategy.StatusFilter(v1.PodRunning),
	)

	// combine into a logic core
	core := strategy.EvaluatorSeive(
		strategy.ApplyFilter(filter),
		strategy.SortCreationTimestampAsc,
		strategy.OlderThen(time.Duration(c.MaxAge)*time.Second),
		strategy.Limit(c.Limit),
	)

	// wrap logic core with Spread strategy
	spread := strategy.Spread(time.Duration(c.MaxAge)*time.Second, core)

	// wrap Spread strategy with cooldown
	return strategy.CoolDown(time.Duration(c.CoolDown)*time.Second, spread)
}
