package maps

type Hashable[K comparable] interface {
	Hash() K
}

func SetFromHash[H Hashable[K], K comparable](m map[K]H, obj H) {
	m[obj.Hash()] = obj
}

func GetFromHash[H Hashable[K], K comparable](m map[K]H, obj H) H {
	return m[obj.Hash()]
}

func ContainFromHash[H Hashable[K], K comparable](m map[K]H, obj H) bool {
	_, contain := m[obj.Hash()]
	return contain
}

// Difference
// In a but not in b
func Difference[K comparable, O any](a map[K]O, b map[K]O) map[K]O {
	diff := map[K]O{}
	for k, v := range a {
		if _, ok := b[k]; !ok {
			diff[k] = v
		}
	}
	return diff
}

func Merge[K comparable, O any](ms ...map[K]O) map[K]O {
	res := map[K]O{}
	for _, m := range ms {
		for k, v := range m {
			res[k] = v
		}
	}
	return res
}

func Keys[K comparable, O any](m map[K]O) []K {
	keys := []K{}
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func Objects[K comparable, O any](m map[K]O) []O {
	objects := []O{}
	for _, v := range m {
		objects = append(objects, v)
	}
	return objects
}

func MapFromHash[H Hashable[K], K comparable](objs []H) map[K]H {
	m := map[K]H{}
	for _, v := range objs {
		m[v.Hash()] = v
	}
	return m
}

func MapFromHashFunc[O any, K comparable](objs []O, hashFunc func(obj O) K) map[K]O {
	m := map[K]O{}
	for _, v := range objs {
		m[hashFunc(v)] = v
	}
	return m
}

func MapSet[K comparable](keys []K) map[K]bool {
	m := map[K]bool{}
	for _, k := range keys {
		m[k] = true
	}
	return m
}
