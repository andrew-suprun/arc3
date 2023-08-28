package renderer

import "math"

type Layout []Constraint

type Constraint struct {
	Size, Flex int
}

func (l Layout) calcSizes(targetSize int) []int {
	result := make([]int, len(l))
	totalSize, totalFlex := 0, 0
	for i, cons := range l {
		result[i] = cons.Size
		totalSize += cons.Size
		totalFlex += cons.Flex
	}
	for totalSize > targetSize {
		idx := 0
		maxSize := result[0]
		for i, size := range result {
			if maxSize < size {
				maxSize = size
				idx = i
			}
		}
		result[idx]--
		totalSize--
	}

	if totalFlex == 0 {
		return result
	}

	if totalSize < targetSize {
		diff := targetSize - totalSize
		remainders := make([]float64, len(l))
		for i, cons := range l {
			rate := float64(diff*cons.Flex) / float64(totalFlex)
			remainders[i] = rate - math.Floor(rate)
			result[i] += int(rate)
		}
		totalSize := 0
		for _, size := range result {
			totalSize += size
		}
		for i := range result {
			if totalSize == targetSize {
				break
			}
			if l[i].Flex > 0 {
				result[i]++
				totalSize++
			}
		}
		for i := range result {
			if totalSize == targetSize {
				break
			}
			if l[i].Flex == 0 {
				result[i]++
				totalSize++
			}
		}
	}

	return result
}
