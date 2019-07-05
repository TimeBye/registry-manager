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
	"regexp"
	"time"

	"github.com/golang/glog"

	"github.com/TimeBye/registry-manager/pkg/utils"
)

type DeleteController struct {
	DryRun       bool          `mapstructure:"dry-run"`
	IntervalHour time.Duration `mapstructure:"interval-hour"`
	MixCount     int           `mapstructure:"mix-count"`
	Tags         struct {
		Include struct {
			Keys      string `mapstructure:"keys"`
			Regex     string `mapstructure:"regex"`
			KeysRegex string
		} `mapstructure:"include"`
		Exclude struct {
			Keys      string `mapstructure:"keys"`
			Regex     string `mapstructure:"regex"`
			KeysRegex string
		} `mapstructure:"exclude"`
	} `mapstructure:"tags"`
}

func (d *DeleteController) NeedDeleteTag(tag Tag, count *int) bool {
	baseTime, _ := time.ParseDuration("1h")
	if *count < d.MixCount {
		*count += 1
		glog.V(2).Infof("跳过 %s ，至少保留: %d 个，当前第: %d 个", tag.Name, d.MixCount, *count)
		return false
	}
	if time.Now().Sub(tag.CreateTime) < d.IntervalHour*baseTime {
		glog.V(2).Infof("跳过 %s ，创建日期在 %d 小时内", tag.Name, d.IntervalHour)
		return false
	}
	match, err := regexp.MatchString(d.Tags.Exclude.Regex, tag.Name)
	utils.CheckErr(err)
	if match {
		glog.V(2).Infof("跳过 %s ，因为匹配排除正则: %s", tag.Name, d.Tags.Exclude.Regex)
		return false
	}
	match, err = regexp.MatchString(d.Tags.Exclude.KeysRegex, tag.Name)
	utils.CheckErr(err)
	if match {
		glog.V(2).Infof("跳过 %s ，因为匹配排除关键字: %s", tag.Name, d.Tags.Exclude.KeysRegex)
		return false
	}
	match, err = regexp.MatchString(d.Tags.Include.Regex, tag.Name)
	utils.CheckErr(err)
	if match {
		glog.V(2).Infof("删除 %s ，因为匹配删除正则: %s", tag.Name, d.Tags.Include.Regex)
		return true
	}
	match, err = regexp.MatchString(d.Tags.Include.KeysRegex, tag.Name)
	utils.CheckErr(err)
	if match {
		glog.V(2).Infof("删除 %s ，因为匹配删除关键字: %s", tag.Name, d.Tags.Include.KeysRegex)
		return true
	}
	return false
}
