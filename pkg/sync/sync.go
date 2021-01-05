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
	"github.com/TimeBye/registry-manager/pkg/global"
	"github.com/TimeBye/registry-manager/pkg/skopeo"
	"github.com/TimeBye/registry-manager/pkg/types"
	"github.com/TimeBye/registry-manager/pkg/utils"
	"github.com/x-mod/glog"
	"strings"
	"sync"
)

func Run() {
	syncPolicy := global.Manager.SyncPolicy
	for _, repository := range syncPolicy.Repositories {
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
	fromTagList := &types.TagList{}
	syncPolicy := global.Manager.SyncPolicy
	fromTagList = skopeo.Tags(global.Manager.SyncPolicy.From, repository)
	glog.V(3).Infof("源库：%s\n获取到 Tag：%v", fromTagList.Repository, fromTagList.Tags)
	if len(tags) == 0 {
		filterFromTagList := syncPolicy.NeedSync(fromTagList.Tags)
		fromTagList.Tags = filterFromTagList
	} else {
		fromTagList.Tags = tags
	}
	glog.Infof("源库：%s\n过滤后 Tag：%v", fromTagList.Repository, fromTagList.Tags)

	toTagList := skopeo.Tags(syncPolicy.To, syncPolicy.ReplaceName(repository))
	glog.Infof("目标库：%s\n获取到 Tag：%v", toTagList.Repository, toTagList.Tags)

	needSyncTags := utils.Difference(fromTagList.Tags, toTagList.Tags)
	glog.Infof("共计需同步 %d 个 Tag：%v", len(needSyncTags), needSyncTags)

	wg := &sync.WaitGroup{}
	for i, tag := range needSyncTags {
		wg.Add(1)
		glog.Infof("当前同步：%d/%d，%s", i+1, len(needSyncTags), tag)
		go skopeo.Copy(repository, tag, wg)
		if (i+1)%global.ProcessLimit == 0 {
			wg.Wait()
		}
	}
	wg.Wait()
}
