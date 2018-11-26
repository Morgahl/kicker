package strategy

import (
	"log"
	"sort"
	"time"

	"k8s.io/api/core/v1"
)

func EvaluatorSeive(evaluators ...Evaluator) Evaluator {
	return func(pods []v1.Pod) []v1.Pod {
		log.Printf("EvaluatorSeive called with %d pods", len(pods))
		for i := range evaluators {
			pods = evaluators[i](pods)
			if len(pods) <= 0 {
				return pods
			}
		}

		log.Printf("EvaluatorSeive exiting with %d pods", len(pods))
		return pods
	}
}

func FilterPodSet(filterSet []v1.Pod) Evaluator {
	return func(pods []v1.Pod) []v1.Pod {
		log.Printf("FilterPodSet called with %d pods", len(pods))
		if len(pods) <= 0 {
			return nil
		}

		out := make([]v1.Pod, 0, len(pods))
		for _, pod := range pods {
			var found bool
			for _, filtPod := range filterSet {
				if pod.Name == filtPod.Name && pod.Namespace == filtPod.Namespace {
					found = true
					break
				}
			}

			if !found {
				out = append(out, pod)
			}
		}

		out = out[:len(out):len(out)]

		log.Printf("FilterPodSet exiting with %d pods", len(out))
		return out
	}
}

// ApplyFilter returns a seive that filters a set of v1.Pods and returns a new slice of pods that match the filter.
func ApplyFilter(filter Filter) Evaluator {
	return func(pods []v1.Pod) []v1.Pod {
		log.Printf("ApplyFilter called with %d pods", len(pods))
		if len(pods) <= 0 {
			return nil
		}

		out := make([]v1.Pod, 0, len(pods))
		for _, pod := range pods {
			if filter(pod) {
				out = append(out, pod)
			}
		}

		out = out[:len(out):len(out)]

		log.Printf("ApplyFilter exiting with %d pods", len(out))
		return out
	}
}

func SortCreationTimestampAsc(pods []v1.Pod) []v1.Pod {
	log.Printf("SortCreationTimestampAsc called with %d pods", len(pods))
	sort.Slice(pods, func(i, j int) bool {
		return pods[i].CreationTimestamp.Time.Before(pods[j].CreationTimestamp.Time)
	})

	log.Printf("SortCreationTimestampAsc exiting with %d pods", len(pods))
	return pods
}

func OlderThan(maxAge time.Duration) Evaluator {
	return func(pods []v1.Pod) []v1.Pod {
		log.Printf("OlderThan called with %d pods", len(pods))
		maxT := time.Now().Add(-maxAge)
		out := make([]v1.Pod, 0, len(pods))
		for i := range pods {
			if pods[i].CreationTimestamp.Time.Before(maxT) {
				out = append(out, pods[i])
			}
		}

		out = out[:len(out):len(out)]

		log.Printf("OlderThan exiting with %d pods", len(out))
		return out
	}
}

func CoolDown(cd time.Duration, seive Evaluator) Evaluator {
	cdWait := time.Time{}
	return func(pods []v1.Pod) []v1.Pod {
		log.Printf("CoolDown called with %d pods", len(pods))
		if time.Now().Before(cdWait) {
			log.Println("CoolDown exiting early due to cool down")
			return nil
		}

		pods = seive(pods)

		if len(pods) > 0 {
			cdWait = time.Now().Add(cd)
		}

		log.Printf("CoolDown exiting with %d pods", len(pods))
		return pods
	}
}

func Spread(maxAge time.Duration, seive Evaluator) Evaluator {
	waitUntil := time.Time{}
	lastCountAtKick := -1
	return func(pods []v1.Pod) []v1.Pod {
		log.Printf("Spread called with %d pods", len(pods))
		startCount := len(pods)
		if time.Now().Before(waitUntil) && startCount <= lastCountAtKick {
			log.Println("Spread exiting early due to spread cool down")
			return nil
		}

		pods = seive(pods)

		if len(pods) > 0 {
			waitUntil = time.Now().Add(maxAge / time.Duration(startCount))
			lastCountAtKick = len(pods)
		}

		log.Printf("Spread exiting with %d pods", len(pods))
		return pods
	}
}

func Limit(limit int64) Evaluator {
	return func(pods []v1.Pod) []v1.Pod {
		log.Printf("Limit called with %d pods", len(pods))
		if int64(len(pods)) > limit {
			pods = pods[:limit:limit]
		}

		log.Printf("Limit exiting with %d pods", len(pods))
		return pods
	}
}
