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

import (
	"fmt"
	"github.com/x-mod/glog"
	"regexp"
	"strings"
)

type SyncPolicy struct {
	FromObj *Registry
	ToObj   *Registry
	From    string `mapstructure:"from"`
	To      string `mapstructure:"to"`
	Start   int    `mapstructure:"start"`
	DryRun  bool   `mapstructure:"dry-run"`
	Replace []struct {
		Old string `mapstructure:"old"`
		New string `mapstructure:"new"`
	} `mapstructure:"replace"`
	Filters []string `mapstructure:"filters"`
}

func (c *SyncPolicy) NeedSync(tags []string) []string {
	t := make([]string, 0)
	for _, k := range tags {
		for _, rule := range c.Filters {
			match, err := regexp.MatchString(rule, k)
			if err != nil {
				glog.Errorf("同步规则：%s，匹配出错：%s", rule, err.Error())
				continue
			}
			if match {
				t = append(t, k)
				continue
			}
		}
	}
	return t
}

func (c *SyncPolicy) ReplaceName(name string) string {
	for _, v := range c.Replace {
		splitName := strings.Split(name, "/")
		countSplitName := len(splitName)
		if countSplitName > 1 && v.Old != "" && strings.Contains(name, v.Old) {
			name = strings.Replace(name, v.Old, v.New, 1)
			break
		} else if countSplitName == 1 && v.Old == "" {
			name = fmt.Sprintf("%s/%s", v.New, name)
			break
		}
	}
	return name
}
