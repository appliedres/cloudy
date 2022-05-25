package cloudy

func ArrayIncludes[T comparable](arr []T, item T) bool {
	for _, me := range arr {
		if item == me {
			return true
		}
	}
	return false
}

// Determines the items that are NOT in this set and returns those
func ArrayDisjoint[T comparable](all []T, arr []T) []T {
	var disjoint []T
	for _, item := range arr {
		if !ArrayIncludes(all, item) {
			disjoint = append(disjoint, item)
		}
	}
	return disjoint
}

func ArrayRemoveAll[T comparable](arr []T, matches func(item T) bool) []T {
	var newarr []T

	for _, item := range arr {
		if !matches(item) {
			newarr = append(newarr, item)
		}
	}
	return newarr
}

// Detemines the items that are in this set
func ArrayIncludesAll[T comparable](all []T, arr []T) []T {
	var includesAll []T
	for _, item := range arr {
		if ArrayIncludes(all, item) {
			includesAll = append(includesAll, item)
		}
	}
	return includesAll
}

func ArrayFindIndex[T comparable](all []T, fn func(item T) bool) int {
	for i, item := range all {
		if fn(item) {
			return i
		}
	}
	return -1
}

func ArrayFirst[T comparable](all []T, fn func(item T) bool) (found T, ok bool) {
	for _, item := range all {
		if fn(item) {
			found = item
			ok = true
			return
		}
	}
	return
}

func ArrayRemoveIndex[T comparable](all []T, i int) []T {
	all[i] = all[len(all)-1]
	return all[:len(all)-1]
}

func ArrayRemove[T comparable](all []T, fn func(item T) bool) ([]T, bool) {
	i := ArrayFindIndex(all, fn)
	if i == -1 {
		return all, false
	}
	all[i] = all[len(all)-1]
	return all[:len(all)-1], true
}
