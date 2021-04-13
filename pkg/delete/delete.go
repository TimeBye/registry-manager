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
	client "github.com/TimeBye/docker-registry-client/registry"
	"github.com/TimeBye/registry-manager/pkg/global"
	"github.com/TimeBye/registry-manager/pkg/skopeo"
	"github.com/TimeBye/registry-manager/pkg/types"
	"github.com/x-mod/glog"
	"strings"
	"sync"
)

var repositories = make([]string, 0)

func Run() {
	deletePolicy := &global.Manager.DeletePolicy
	deletePolicy.Init()
	for _, reg := range deletePolicy.Registries {
		registry, ok := global.Manager.Registries[reg]
		if !ok {
			glog.Exitf("未在 registries 中找到仓库: %s", reg)
		}
		deletePolicy.RegistriesObj = append(deletePolicy.RegistriesObj, registry)
	}
	for _, registry := range deletePolicy.RegistriesObj {
		if len(registry.Repositories) > 0 {
			repositories = registry.Repositories
		} else {
			registryClient := &client.Registry{}
			if !registry.Insecure {
				registryClient, _ = client.New(registry.Url, registry.Username, registry.Password)
			} else {
				registryClient, _ = client.NewInsecure(registry.Url, registry.Username, registry.Password)
			}
			var err error
			repositories, err = registryClient.Repositories()
			if err != nil {
				glog.Exitf("获取仓库出错：%s", err.Error())
			}
		}
		deleteTags(registry)
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

func deleteTags(registry types.Registry) {
	deletePolicy := &global.Manager.DeletePolicy
	repositoriesCount := len(repositories)
	glog.Infof("获取到仓库数量：%d", repositoriesCount)
	for startIndex := deletePolicy.Start; startIndex < repositoriesCount; startIndex++ {
		repository := repositories[startIndex]
		glog.Infof("当前处理第 %d/%d 个仓库：%s", startIndex+1, repositoriesCount, repository)
		wg := &sync.WaitGroup{}
		// 如果指定了tag，则直接删除
		if strings.Contains(repository, ":") {
			tag := strings.Split(repository, ":")[1]
			glog.Infof("删除镜像: docker://%s/%s", registry.Uri.Host, repository)
			if !deletePolicy.DryRun {
				wg.Add(1)
				skopeo.Delete(registry, repository, tag, wg)
			}
			wg.Wait()
		} else {
			// 否则获取该镜像的所有tag进行判断后删除
			tags := skopeo.Tags(registry, repository)
			tagsTotal := len(tags.Tags)
			glog.Infof("仓库：%s，所有 Tag 总数：%d，%+v", tags.Repository, tagsTotal, tags.Tags)
			if tagsTotal <= deletePolicy.MixCount {
				continue
			}
			needKeepTags, needDeleteTags, noSemVerTags := deletePolicy.AnalysisTags(tags.Tags)
			glog.Infof("仓库：%s，需保留的 Tag 总数：%d，%+v", tags.Repository, len(needKeepTags), needKeepTags)

			if len(noSemVerTags) > 0 {
				glog.Infof("仓库：%s，非语义化 Tag 总数：%d，%+v", tags.Repository, len(noSemVerTags), noSemVerTags)
				if deletePolicy.SemVer {
					for tagIndex, tag := range noSemVerTags {
						glog.Infof("删除镜像：%s:%s", tags.Repository, tag)
						if !deletePolicy.DryRun {
							wg.Add(1)
							go skopeo.Delete(registry, repository, tag, wg)
						}
						if (tagIndex+1)%global.ProcessLimit == 0 {
							wg.Wait()
						}
					}
				}
			}

			count := len(needDeleteTags) - deletePolicy.MixCount
			if count > 0 {
				glog.Infof("仓库：%s，需删除的 Tag 总数：%d，%+v", tags.Repository, count, needDeleteTags[:count])
				for tagIndex, tag := range needDeleteTags[:count] {
					glog.Infof("删除镜像: %s:%s", tags.Repository, tag)
					if !deletePolicy.DryRun {
						wg.Add(1)
						go skopeo.Delete(registry, repository, tag, wg)
					}
					if (tagIndex+1)%global.ProcessLimit == 0 {
						wg.Wait()
					}
				}
			}
			wg.Wait()
		}
	}
}
