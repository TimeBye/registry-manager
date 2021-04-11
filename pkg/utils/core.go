// Copyright © 2019 TimeBye zhongziling@vip.qq.com
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package utils

func Difference(sliceA []string, sliceB []string) []string {
	diff := make([]string, 0)
	diffMap := make(map[string]int)

	for _, v := range sliceA {
		diffMap[v] = 1
	}
	for _, v := range sliceB {
		diffMap[v] = diffMap[v] - 1
	}

	for k, v := range diffMap {
		if v > 0 {
			diff = append(diff, k)
		}
	}
	return diff
}
