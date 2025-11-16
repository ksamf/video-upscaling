package utils

func ClosestStandardHeight(height int) int {
	closest := StandardHeights[0]
	minDiff := Abs(height - closest)
	for _, h := range StandardHeights[1:] {
		diff := Abs(height - h)
		if diff < minDiff {
			minDiff = diff
			closest = h
		}
	}
	return closest
}

func Abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
func LowerStandardRes(base int) []int {
	if base <= 144 {
		return []int{144}
	}
	var result []int
	for _, h := range StandardHeights {
		if h < base {
			result = append(result, h)
		}
	}
	return result
}
