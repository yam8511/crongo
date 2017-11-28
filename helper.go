package crongo

// 如同 nodejs 的 indexOf
// 從 arr 裡面找出 val 的位子
func indexOf(arr []int, val int) int {
	for index, value := range arr {
		if value == val {
			return index
		}
	}
	return -1
}
