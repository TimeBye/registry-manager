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

package types

import "time"

type Tags []Tag

func (tags Tags) Len() int {
	return len(tags)
}
func (tags Tags) Swap(i, j int) {
	tags[i], tags[j] = tags[j], tags[i]
}
func (tags Tags) Less(i, j int) bool {
	return tags[j].CreateTime.Before(tags[i].CreateTime)
}

type Tag struct {
	Name       string
	CreateTime time.Time
}
