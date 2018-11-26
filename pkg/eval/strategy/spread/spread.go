package spread

import (
	"log"
	"time"

	"github.com/curlymon/kicker/pkg/conf"
	"github.com/curlymon/kicker/pkg/eval"
	"k8s.io/api/core/v1"
)

func init() {
	if err := eval.RegisterEvaluatorConstructor(conf.StrategySpread, Spread); err != nil {
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
func Spread(c conf.Criteria) (eval.Evaluator, error) {
	filter := eval.And(
		eval.NamePrefixFilter(c.Name),
		eval.NameSpaceFilter(c.Namespace),
		eval.StatusFilter(v1.PodRunning),
	)

	max := time.Second * time.Duration(c.MaxAge)

	waitUntil := time.Time{}
	cd := time.Second * time.Duration(c.CoolDown)
	cdWait := time.Time{}
	lastCountAtKick := -1

	return func(pods []v1.Pod) []v1.Pod {
		if len(pods) <= 0 {
			return nil
		}

		now := time.Now()
		if now.Before(cdWait) {
			return nil
		}

		pods = filter.Apply(pods)
		if len(pods) <= 0 {
			return nil
		}

		// we do not want to fire this off too quickly to keep things spread out
		if now.Before(waitUntil) && len(pods) <= lastCountAtKick {
			return nil
		}

		eval.SortCreationTimestampAsc(pods)

		maxT := now.Add(-max)
		out := make([]v1.Pod, 0, c.Limit)
		for i := range pods {
			if c.Limit <= int64(len(out)) {
				break
			}

			if pods[0].CreationTimestamp.Time.Before(maxT) {
				out = append(out, pods[i])
			}

			break // due to sorting above this becomes an early exit
		}

		if len(out) > 0 {
			cdWait = now.Add(cd)
			waitUntil = now.Add(max / time.Duration(len(pods)))
			lastCountAtKick = len(pods)
		}

		return out
	}, nil
}
