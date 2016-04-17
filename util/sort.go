package util

type ByInt64 []int64

func (e ByInt64) Len() int           { return len(e) }
func (e ByInt64) Swap(i, j int)      { e[i], e[j] = e[j], e[i] }
func (e ByInt64) Less(i, j int) bool { return e[i] < e[j] }
