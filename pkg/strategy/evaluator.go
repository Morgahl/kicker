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

// ApplyFilter returns a eval that filters a set of v1.Pods and returns a new slice of pods that match the filter.
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

func CoolDown(cd time.Duration, eval Evaluator) Evaluator {
	cdWait := time.Time{}
	return func(pods []v1.Pod) []v1.Pod {
		log.Printf("CoolDown called with %d pods", len(pods))
		if time.Now().Before(cdWait) {
			log.Println("CoolDown exiting early due to cool down")
			return nil
		}

		pods = eval(pods)

		if len(pods) > 0 {
			log.Printf("CoolDown setting cool down for %s", cd)
			cdWait = time.Now().Add(cd)
		}

		log.Printf("CoolDown exiting with %d pods", len(pods))
		return pods
	}
}

func Spread(maxAge time.Duration, eval Evaluator) Evaluator {
	waitUntil := time.Time{}
	return func(pods []v1.Pod) []v1.Pod {
		log.Printf("Spread called with %d pods", len(pods))
		if time.Now().Before(waitUntil) {
			log.Println("Spread exiting early due to spread cool down")
			return nil
		}

		maxT := maxAge / time.Duration(len(pods))
		pods = eval(pods)

		if len(pods) > 0 {
			log.Printf("Spread setting cool down for %s", maxT)
			waitUntil = time.Now().Add(maxT)
		}

		log.Printf("Spread exiting with %d pods", len(pods))
		return pods
	}
}

func SpreadFast(maxAge time.Duration, limit int64, eval Evaluator) Evaluator {
	lastEvict := time.Time{}
	return func(pods []v1.Pod) []v1.Pod {
		log.Printf("SpreadFast called with %d pods", len(pods))
		minAge := maxAge / time.Duration(len(pods))
		if time.Now().Before(lastEvict.Add(minAge)) {
			log.Println("SpreadFast exiting early due to spread cool down")
			return nil
		}

		pods = eval(pods)
		pods = OlderThan(minAge)(pods) // this prevents this from firing as soon as this strategy is first run, unless we actually HAVE a pod elidgeable.
		pods = Limit(limit)(pods)

		if len(pods) > 0 {
			log.Printf("SpreadFast setting cool down for %s", minAge)
			lastEvict = time.Now()
		}

		log.Printf("SpreadFast exiting with %d pods", len(pods))
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
