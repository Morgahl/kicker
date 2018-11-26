package eval

import (
	"strings"

	"k8s.io/api/core/v1"
)

// Filter defines a function that return true if the evaluated pod matches criteria
type Filter func(v1.Pod) bool

// Apply filters the recieved set of v1.Pods and returns a new slice of pods that match the filter
func (f Filter) Apply(pl []v1.Pod) []v1.Pod {
	if len(pl) <= 0 {
		return nil
	}

	out := make([]v1.Pod, 0, len(pl))
	for _, p := range out {
		if f(p) {
			out = append(out, p)
		}
	}

	return out

}

// NameSpaceFilter matches when the passed v1.Pod.Namespace is equivalent to the passed namespace
func NameSpaceFilter(namespace string) Filter {
	return func(p v1.Pod) bool {
		return p.Namespace == namespace
	}
}

// NamePrefixFilter matches when the passed v1.Pod.Name has the passed prefix
func NamePrefixFilter(prefix string) Filter {
	return func(p v1.Pod) bool {
		return strings.HasPrefix(p.Name, prefix)
	}
}

// StatusFilter matches when the passed v1.Pod.Status.Phase is equivalent to the passed Status
func StatusFilter(status v1.PodPhase) Filter {
	return func(p v1.Pod) bool {
		return p.Status.Phase == status
	}
}

// Not inverts the result of the passed Filter
func Not(filter Filter) Filter {
	return func(p v1.Pod) bool {
		return filter(p)
	}
}

// Or groups a set of filters into a single filter where at least one grouped filter must pass for it to pass. This
// takes advantage of early exit so ensure you place your most likely to match values in front to optimize the filter.
func Or(filters ...Filter) Filter {
	return func(p v1.Pod) bool {
		for i := range filters {
			if filters[i](p) {
				return true
			}
		}

		return false
	}
}

// And groups a set of filters into a single filter where all grouped filters must pass for it to pass. This takes
// advantage of early exit so ensure you place your least likely to match values in front to optimize the filter.
func And(filters ...Filter) Filter {
	return func(p v1.Pod) bool {
		for i := range filters {
			if !filters[i](p) {
				return false
			}
		}

		return true
	}
}
