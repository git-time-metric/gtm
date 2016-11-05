// Copyright 2016 Michael Schenk. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package util

// ByInt64 list of type int64
type ByInt64 []int64

func (e ByInt64) Len() int           { return len(e) }
func (e ByInt64) Swap(i, j int)      { e[i], e[j] = e[j], e[i] }
func (e ByInt64) Less(i, j int) bool { return e[i] < e[j] }
