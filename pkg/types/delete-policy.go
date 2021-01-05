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
	"regexp"
	"sort"
	"strings"

	"github.com/hashicorp/go-version"
	"github.com/x-mod/glog"

	"github.com/TimeBye/registry-manager/pkg/utils"
)

type DeletePolicy struct {
	Registries   []string `mapstructure:"registries"`
	Repositories []string `mapstructure:"repositories"`
	Start        int      `mapstructure:"start"`
	DryRun       bool     `mapstructure:"dry-run"`
	SemVer       bool     `mapstructure:"sem-ver"`
	MixCount     int      `mapstructure:"mix-count"`
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

// 通过关键字生成正则表达式
func generateRegexByKeys(keys string) string {
	if strings.HasPrefix(keys, ",") {
		keys = keys[1:]
	}
	if strings.HasSuffix(keys, ",") {
		keys = keys[:len(keys)-1]
	}
	if len(keys) == 0 || len(strings.Replace(keys, ",", "", -1)) == 0 {
		return ":"
	}
	return fmt.Sprintf(".*%s.*", strings.Replace(keys, ",", ".*|.*", -1))
}

// 初始化 registry 客户端及检验参数
func (d *DeletePolicy) Init() {
	// 根据关键字生成正则表达式
	d.Tags.Include.KeysRegex = generateRegexByKeys(d.Tags.Include.Keys)
	d.Tags.Exclude.KeysRegex = generateRegexByKeys(d.Tags.Exclude.Keys)

	// 值为空时设置为无法匹配到的字符
	if len(d.Tags.Include.Regex) == 0 {
		d.Tags.Include.Regex = ":"
	}
	if len(d.Tags.Include.KeysRegex) == 0 {
		d.Tags.Include.KeysRegex = ":"
	}
	if len(d.Tags.Exclude.Regex) == 0 {
		d.Tags.Exclude.Regex = ":"
	}
	if len(d.Tags.Exclude.KeysRegex) == 0 {
		d.Tags.Exclude.KeysRegex = ":"
	}
}

func (d *DeletePolicy) AnalysisTags(tags []string) (needKeepTags, needDeleteTags, noSemVerTags []string) {
	needKeepTags = make([]string, 0)
	needDeleteTags = make([]string, 0)
	noSemVerTags = make([]string, 0)
	for _, tag := range tags {
		if _, err := version.NewVersion(tag); err != nil {
			noSemVerTags = append(noSemVerTags, tag)
			continue
		}
		if d.NeedDelete(tag) {
			needDeleteTags = append(needDeleteTags, tag)
		} else {
			needKeepTags = append(needKeepTags, tag)
		}
	}

	// 给需要删除 tag 进行排序
	needDeleteTagsVersion := make([]*version.Version, 0)
	for _, tag := range needDeleteTags {
		v, _ := version.NewVersion(tag)
		needDeleteTagsVersion = append(needDeleteTagsVersion, v)
	}
	sort.Sort(version.Collection(needDeleteTagsVersion))
	for i, tag := range needDeleteTagsVersion {
		needDeleteTags[i] = tag.String()
	}

	return
}

func (d *DeletePolicy) NeedDelete(tag string) bool {
	match, err := regexp.MatchString(d.Tags.Exclude.Regex, tag)
	utils.CheckErr(err)
	if match {
		glog.V(2).Infof("跳过 %s ，因为匹配排除正则: %s", tag, d.Tags.Exclude.Regex)
		return false
	}
	match, err = regexp.MatchString(d.Tags.Exclude.KeysRegex, tag)
	utils.CheckErr(err)
	if match {
		glog.V(2).Infof("跳过 %s ，因为匹配排除关键字: %s", tag, d.Tags.Exclude.KeysRegex)
		return false
	}
	match, err = regexp.MatchString(d.Tags.Include.Regex, tag)
	utils.CheckErr(err)
	if match {
		glog.V(2).Infof("删除 %s ，因为匹配删除正则: %s", tag, d.Tags.Include.Regex)
		return true
	}
	match, err = regexp.MatchString(d.Tags.Include.KeysRegex, tag)
	utils.CheckErr(err)
	if match {
		glog.V(2).Infof("删除 %s ，因为匹配删除关键字: %s", tag, d.Tags.Include.KeysRegex)
		return true
	}
	return false
}
