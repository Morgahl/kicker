package eval

import (
	"sort"

	"k8s.io/api/core/v1"
)

// SortCreationTimestampAsc sort the passed pods in place in ascending order
func SortCreationTimestampAsc(pods []v1.Pod) {
	sort.Slice(pods, func(i, j int) bool {
		return pods[i].CreationTimestamp.Time.Before(pods[j].CreationTimestamp.Time)
	})
}
