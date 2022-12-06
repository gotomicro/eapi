package utils

import "sort"

func RangeMapInOrder[K comparable, V any](values map[K]V, sorter func(i, j K) bool, visitor func(K, V)) {
	var keys []K
	for k, _ := range values {
		keys = append(keys, k)
	}

	sort.Slice(keys, func(i, j int) bool { return sorter(keys[i], keys[j]) })

	for _, key := range keys {
		visitor(key, values[key])
	}
}
