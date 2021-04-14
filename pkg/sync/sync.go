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

package sync

import (
	"fmt"
	client "github.com/TimeBye/docker-registry-client/registry"
	"github.com/TimeBye/registry-manager/pkg/global"
	"github.com/TimeBye/registry-manager/pkg/skopeo"
	"github.com/TimeBye/registry-manager/pkg/types"
	"github.com/TimeBye/registry-manager/pkg/utils"
	"github.com/x-mod/glog"
	"strings"
	"sync"
)

var repositories = make([]string, 0)

func Run() {
	syncPolicy := &global.Manager.SyncPolicy
	var ok bool
	syncPolicy.FromObj, ok = global.Manager.Registries[syncPolicy.From]
	if !ok {
		glog.Exitf("未在 registries 中找到 from 仓库: %s", syncPolicy.From)
	}
	syncPolicy.ToObj, ok = global.Manager.Registries[syncPolicy.To]
	if !ok {
		glog.Exitf("未在 registries 中找到 to 仓库: %s", syncPolicy.To)
	}

	from := syncPolicy.FromObj
	if len(from.Repositories) > 0 {
		repositories = from.Repositories
	} else {
		registryClient := &client.Registry{}
		if !from.Insecure {
			registryClient, _ = client.New(from.Url, from.Username, from.Password)
		} else {
			registryClient, _ = client.NewInsecure(from.Url, from.Username, from.Password)
		}
		var err error
		repositories, err = registryClient.Repositories()
		if err != nil {
			glog.Exitf("获取仓库出错：%s", err.Error())
		}
	}

	repositoriesCount := len(repositories)
	glog.Infof("获取到仓库数量：%d", repositoriesCount)
	for startIndex := syncPolicy.Start; startIndex < repositoriesCount; startIndex++ {
		repository := repositories[startIndex]
		glog.Infof("当前处理第 %d/%d 个仓库：%s", startIndex+1, repositoriesCount, repository)
		repositoryAndTag := strings.Split(repository, ":")
		if len(repositoryAndTag) == 1 {
			syncAll(repository)
		} else if len(repositoryAndTag) == 2 {
			syncAll(repositoryAndTag[0], repositoryAndTag[1])
		} else {
			glog.Exitf("镜像地址错误：%s", repository)
		}
	}

	if len(global.FailedList) > 0 {
		glog.Exitf("同步失败，列表如下：%s",
			func() string {
				msg := ""
				for k, v := range global.FailedList {
					msg = fmt.Sprintf("%s\n%s %s", msg, k, v)
				}
				return msg
			}())
	} else {
		glog.Info("同步镜像成功")
	}
}

func syncAll(repository string, tags ...string) {
	syncPolicy := &global.Manager.SyncPolicy
	fromTagList := &types.TagList{}
	fromTagList = skopeo.Tags(syncPolicy.FromObj, repository)
	glog.V(3).Infof("源库：%s\n获取到 Tag：%v", fromTagList.Repository, fromTagList.Tags)
	if len(tags) == 0 {
		filterFromTagList := syncPolicy.NeedSync(fromTagList.Tags)
		fromTagList.Tags = filterFromTagList
	} else {
		fromTagList.Tags = tags
	}
	glog.Infof("源库：%s\n过滤后 Tag：%v", fromTagList.Repository, fromTagList.Tags)

	var toTagList *types.TagList
	if syncPolicy.Force {
		toTagList = &types.TagList{
			Repository: fmt.Sprintf("%s/%s",
				syncPolicy.ToObj.Uri.Host, syncPolicy.ReplaceName(repository)),
			Tags: []string{},
		}
	} else {
		toTagList = skopeo.Tags(syncPolicy.ToObj, syncPolicy.ReplaceName(repository))
		glog.Infof("目标库：%s\n获取到 Tag：%v", toTagList.Repository, toTagList.Tags)
	}

	needSyncTags := utils.Difference(fromTagList.Tags, toTagList.Tags)
	glog.Infof("共计需同步 %d 个 Tag：%v", len(needSyncTags), needSyncTags)

	wg := &sync.WaitGroup{}
	for i, tag := range needSyncTags {
		glog.Infof("当前同步：%d/%d，%s", i+1, len(needSyncTags), tag)
		if !syncPolicy.DryRun {
			wg.Add(1)
			go skopeo.Copy(repository, tag, wg)
			if (i+1)%global.ProcessLimit == 0 {
				wg.Wait()
			}
		}
	}
	wg.Wait()
}
