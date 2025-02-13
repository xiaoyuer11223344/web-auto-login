package utils

import "github.com/go-rod/rod"

func GetMaxKey(formsScore map[*rod.Element]int) (*rod.Element, int, bool) {
	if len(formsScore) == 0 {
		return nil, 0, false // 如果 map 为空，返回 nil 和 false
	}

	var maxElement *rod.Element
	maxValue := -1 << 63 // 初始化为最小整数值

	for key, value := range formsScore {
		if value > maxValue {
			maxValue = value
			maxElement = key
		}
	}

	return maxElement, maxValue, true
}
