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
	"github.com/golang/glog"

	. "github.com/TimeBye/registry-manager/pkg/types"
)

func Run() {
	repositories := Manager.Repositories()
	repositoriesCount := len(repositories)
	glog.Infof("获取到仓库数量：%d", repositoriesCount)
	for i, repository := range repositories {
		glog.Infof("当前处理第 %d/%d 个仓库: %s", i+1, repositoriesCount, repository)
		tags := Manager.Tags(repository)
		if len(tags) <= Manager.DeleteController.MixCount {
			continue
		}
		tagObjs := Manager.TagObjs(repository)
		var count = 0

		for _, tag := range tagObjs {
			if Manager.DeleteController.NeedDeleteTag(tag, &count) {
				glog.Infof("删除镜像: %s:%s", repository, tag.Name)
				if !Manager.DeleteController.DryRun {
					Manager.DeleteManifest(repository, tag.Name)
				}
			}
		}

	}
}
