package strategy

import (
	"sort"
	"time"

	"k8s.io/api/core/v1"
)

func EvaluatorSeive(evaluators ...Evaluator) Evaluator {
	return func(pods []v1.Pod) []v1.Pod {
		for i := range evaluators {
			pods = evaluators[i](pods)
			if len(pods) <= 0 {
				return pods
			}
		}

		return pods
	}
}

// ApplyFilter returns a seive that filters a set of v1.Pods and returns a new slice of pods that match the filter.
func ApplyFilter(filter Filter) Evaluator {
	return func(pl []v1.Pod) []v1.Pod {
		if len(pl) <= 0 {
			return nil
		}

		out := make([]v1.Pod, 0, len(pl))
		for _, p := range out {
			if filter(p) {
				out = append(out, p)
			}
		}

		return out
	}
}

func SortCreationTimestampAsc(pods []v1.Pod) []v1.Pod {
	sort.Slice(pods, func(i, j int) bool {
		return pods[i].CreationTimestamp.Time.Before(pods[j].CreationTimestamp.Time)
	})

	return pods
}

func OlderThen(maxAge time.Duration) Evaluator {
	return func(pods []v1.Pod) []v1.Pod {
		maxT := time.Now().Add(maxAge)
		out := make([]v1.Pod, 0, len(pods))
		for i := range pods {
			if pods[i].CreationTimestamp.Time.Before(maxT) {
				out = append(out, pods[i])
			}
		}

		return out[:len(out):len(out)]
	}
}

func CoolDown(cd time.Duration, seive Evaluator) Evaluator {
	cdWait := time.Time{}
	return func(pods []v1.Pod) []v1.Pod {
		if time.Now().Before(cdWait) {
			return nil
		}

		pods = seive(pods)

		if len(pods) > 0 {
			cdWait = time.Now().Add(cd)
		}

		return pods
	}
}

func Spread(maxAge time.Duration, seive Evaluator) Evaluator {
	waitUntil := time.Time{}
	lastCountAtKick := -1
	return func(pods []v1.Pod) []v1.Pod {
		startCount := len(pods)
		if time.Now().Before(waitUntil) && startCount <= lastCountAtKick {
			return nil
		}

		pods = seive(pods)

		if len(pods) > 0 {
			waitUntil = time.Now().Add(maxAge / time.Duration(startCount))
			lastCountAtKick = len(pods)
		}

		return pods
	}
}

func Limit(limit int64) Evaluator {
	return func(pods []v1.Pod) []v1.Pod {
		if int64(len(pods)) > limit {
			return pods[:limit:limit]
		}

		return pods
	}
}
