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
	"sort"
	"strings"
	"time"

	"github.com/docker/distribution/manifest/schema1"
	"github.com/golang/glog"
	"github.com/tidwall/gjson"

	"github.com/TimeBye/docker-registry-client/registry"
	"github.com/TimeBye/registry-manager/pkg/utils"
)

type RegistryManager struct {
	Server                string           `mapstructure:"server"`
	InsecureSkipTLSVerify bool             `mapstructure:"insecure-skip-tls-verify"`
	Username              string           `mapstructure:"username"`
	Password              string           `mapstructure:"password"`
	DeleteController      DeleteController `mapstructure:"delete-policy"`
	Client                *registry.Registry
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
func (r *RegistryManager) Init() {
	var err error
	if r.InsecureSkipTLSVerify {
		r.Client, err = registry.New(r.Server, r.Username, r.Password)
	} else {
		r.Client, err = registry.NewInsecure(r.Server, r.Username, r.Password)
	}
	if err != nil {
		glog.Exit(err)
	}

	// 根据关键字生成正则表达式
	r.DeleteController.Tags.Include.KeysRegex = generateRegexByKeys(r.DeleteController.Tags.Include.Keys)
	r.DeleteController.Tags.Exclude.KeysRegex = generateRegexByKeys(r.DeleteController.Tags.Exclude.Keys)

	// 值为空时设置为无法匹配到的字符
	if len(r.DeleteController.Tags.Include.Regex) == 0 {
		r.DeleteController.Tags.Include.Regex = ":"
	}
	if len(r.DeleteController.Tags.Include.KeysRegex) == 0 {
		r.DeleteController.Tags.Include.KeysRegex = ":"
	}
	if len(r.DeleteController.Tags.Exclude.Regex) == 0 {
		r.DeleteController.Tags.Exclude.Regex = ":"
	}
	if len(r.DeleteController.Tags.Exclude.KeysRegex) == 0 {
		r.DeleteController.Tags.Exclude.KeysRegex = ":"
	}
}

// 获取仓库列表
func (r *RegistryManager) Repositories() []string {
	repositories, e := r.Client.Repositories()
	utils.CheckErr(e)
	return repositories
}

// 获取镜像层信息
func (r *RegistryManager) Manifest(repository, reference string) *schema1.SignedManifest {
	signedManifest, e := r.Client.Manifest(repository, reference)
	utils.CheckErr(e)
	return signedManifest
}

// 通过镜像层信息解析镜像创建时间
func (r *RegistryManager) ManifestCreateTime(s *schema1.SignedManifest) time.Time {
	var createdTime time.Time

	for _, v := range s.History {
		t := gjson.Get(v.V1Compatibility, "created").Time()
		if createdTime.Before(t) {
			createdTime = t
		}
	}

	return createdTime
}

// 获取tag列表
func (r *RegistryManager) Tags(repository string) []string {
	tags, e := r.Client.Tags(repository)
	utils.CheckErr(e)
	return tags
}

// 获取带创建时间了的tag列表
func (r *RegistryManager) TagObjs(repository string) []Tag {
	tags, e := r.Client.Tags(repository)
	utils.CheckErr(e)

	// 获取镜像生成时间
	tagsObj := make([]Tag, 0)

	for _, tag := range tags {
		if signedManifest := r.Manifest(repository, tag); signedManifest != nil {
			t := r.ManifestCreateTime(signedManifest)
			tagsObj = append(tagsObj, Tag{
				Name:       tag,
				CreateTime: t,
			})
		}
	}

	sort.Sort(Tags(tagsObj))
	return tagsObj
}

func (r *RegistryManager) DeleteManifest(repository, tag string) {
	digest, e := r.Client.ManifestDigestV2(repository, tag)
	utils.CheckErr(e)
	r.Client.DeleteManifest(repository, digest)
	utils.CheckErr(e)
}
