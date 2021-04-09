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

package delete

import (
	"fmt"
	"github.com/TimeBye/docker-registry-client/registry"
	"github.com/TimeBye/registry-manager/pkg/global"
	"github.com/TimeBye/registry-manager/pkg/skopeo"
	"github.com/TimeBye/registry-manager/pkg/utils"
	"github.com/x-mod/glog"
	"sync"
)

var repositories = make([]string, 0)

func Run() {
	deletePolicy := &global.Manager.DeletePolicy
	deletePolicy.Init()
	for _, reg := range deletePolicy.Registries {
		r := global.Manager.Registries[reg]
		if len(deletePolicy.Repositories) == 0 {
			registryClient := &registry.Registry{}
			if !r.Insecure {
				registryClient, _ = registry.New(r.Url, r.Username, r.Password)
			} else {
				registryClient, _ = registry.NewInsecure(r.Url, r.Username, r.Password)
			}
			var err error
			repositories, err = registryClient.Repositories()
			utils.CheckErr(err)
		} else {
			repositories = deletePolicy.Repositories
		}
		deleteTags(reg)
	}
	if len(global.FailedList) > 0 {
		glog.Exitf("有删除失败的镜像，列表如下：%s",
			func() string {
				msg := ""
				for k, _ := range global.FailedList {
					msg = fmt.Sprintf("%s\n%s", msg, k)
				}
				return msg
			}())
	} else {
		glog.Info("删除镜像成功")
	}
}

func deleteTags(r string) {
	deletePolicy := &global.Manager.DeletePolicy
	repositoriesCount := len(repositories)
	glog.Infof("获取到仓库数量：%d", repositoriesCount)
	for i := deletePolicy.Start; i < repositoriesCount; i++ {
		glog.Infof("当前处理第 %d/%d 个仓库：%s", i+1, repositoriesCount, repositories[i])
		tags := skopeo.Tags(r, repositories[i])
		tagsTotal := len(tags.Tags)
		glog.Infof("仓库：%s，所有 Tag 总数：%d，%+v", tags.Repository, tagsTotal, tags.Tags)
		if tagsTotal <= deletePolicy.MixCount {
			continue
		}
		needKeepTags, needDeleteTags, noSemVerTags := deletePolicy.AnalysisTags(tags.Tags)
		glog.Infof("仓库：%s，需保留的 Tag 总数：%d，%+v", tags.Repository, len(needKeepTags), needKeepTags)

		wg := &sync.WaitGroup{}
		if len(noSemVerTags) > 0 {
			glog.Infof("仓库：%s，非语义化 Tag 总数：%d，%+v", tags.Repository, len(noSemVerTags), noSemVerTags)
			if deletePolicy.SemVer {
				for _, tag := range noSemVerTags {
					wg.Add(1)
					glog.Infof("删除镜像：%s:%s", tags.Repository, tag)
					if !deletePolicy.DryRun {
						go skopeo.Delete(r, repositories[i], tag, wg)
					}
					if (i+1)%global.ProcessLimit == 0 {
						wg.Wait()
					}
				}
			}
		}

		count := len(needDeleteTags) - deletePolicy.MixCount
		if count > 0 {
			glog.Infof("仓库：%s，需删除的 Tag 总数：%d，%+v", tags.Repository, count, needDeleteTags[:count])
			for _, tag := range needDeleteTags[:count] {
				wg.Add(1)
				glog.Infof("删除镜像: %s:%s", tags.Repository, tag)
				if !deletePolicy.DryRun {
					go skopeo.Delete(r, repositories[i], tag, wg)
				}
				if (i+1)%global.ProcessLimit == 0 {
					wg.Wait()
				}
			}
		}
		wg.Wait()
	}
}
