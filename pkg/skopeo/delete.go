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

package skopeo

import (
	"fmt"
	"github.com/TimeBye/registry-manager/pkg/global"
	"github.com/TimeBye/registry-manager/pkg/types"
	"github.com/go-cmd/cmd"
	"github.com/x-mod/glog"
	"strings"
	"sync"
)

func Delete(registry types.Registry, repository, tag string, wg *sync.WaitGroup) {
	defer wg.Done()
	retryStart := 1
	tagDeleteArgs := generateDeleteTagArgs(registry, repository, tag)
RePlay:
	if retryStart > global.Retry {
		global.FailedList[tagDeleteArgs[len(tagDeleteArgs)-1]] = ""
		glog.Errorf("删除镜像已达最大重试次数：%d，%s", global.Retry, tagDeleteArgs[len(tagDeleteArgs)-1])
	} else {
		tagDeleteCmd := cmd.NewCmd("skopeo", tagDeleteArgs...)
		result := <-tagDeleteCmd.Start()
		if result.Exit != 0 {
			glog.Errorf("删除 Tag 出错：%s\n错误信息：%s",
				tagDeleteArgs[len(tagDeleteArgs)-1], strings.Join(result.Stderr, ""))
			retryStart = retryStart + 1
			goto RePlay
		}
	}
}

func generateDeleteTagArgs(registry types.Registry, repository, tag string) []string {
	cmd := make([]string, 0)
	cmd = append(cmd, "delete")
	if registry.Insecure {
		cmd = append(cmd, "--tls-verify=false")
	}
	if registry.Username != "" && registry.Password != "" {
		cmd = append(cmd, fmt.Sprintf("--creds=%s:%s", registry.Username, registry.Password))
	}
	cmd = append(cmd, fmt.Sprintf("docker://%s/%s:%s", registry.Uri.Host, repository, tag))
	return cmd
}
