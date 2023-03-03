package kit

func SliceContain[T comparable](elems []T, target T) bool {
	for _, v := range elems {
		if v == target {
			return true
		}
	}
	return false
}

func SliceRemoveFirst[T comparable](l []T, item T) []T {
	for i, other := range l {
		if other == item {
			return append(l[:i], l[i+1:]...)
		}
	}
	return l
}
