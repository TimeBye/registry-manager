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
	"encoding/json"
	"fmt"
	"github.com/TimeBye/registry-manager/pkg/global"
	"github.com/TimeBye/registry-manager/pkg/types"
	"github.com/go-cmd/cmd"
	"github.com/x-mod/glog"
	"net/url"
	"strings"
)

func Tags(registry, repository string) *types.TagList {
	retryStart := 1
	tagList := &types.TagList{}
	tagListArgs := generateListTagArgs(registry, repository)
RePlay:
	if retryStart > global.Retry {
		tagList.Repository = tagListArgs[len(tagListArgs)-1]
		glog.Errorf("获取 Tag 已超过最大重试次数：%d，返回空白列表", global.Retry)
	} else {
		tagListCmd := cmd.NewCmd("skopeo", tagListArgs...)
		result := <-tagListCmd.Start()
		if result.Exit != 0 {
			glog.Errorf("获取 Tag 出错：%s\n错误信息：%s",
				tagListArgs[len(tagListArgs)-1], strings.Join(result.Stderr, ""))
			retryStart = retryStart + 1
			goto RePlay
		}
		if err := json.Unmarshal([]byte(strings.Join(result.Stdout, "")), tagList); err != nil {
			glog.Errorf("解析 Tag 出错：%s\n错误信息：%s",
				tagListArgs[len(tagListArgs)-1], err.Error())
		}
	}
	return tagList
}

func generateListTagArgs(registry, repository string) []string {
	cmd := make([]string, 0)
	cmd = append(cmd, "list-tags")
	r := global.Manager.Registries[registry]
	if r.Username != "" && r.Password != "" {
		cmd = append(cmd, "--creds")
		cmd = append(cmd, fmt.Sprintf("%s:%s", r.Username, r.Password))
	}
	if r.Insecure {
		cmd = append(cmd, "--tls-verify")
		cmd = append(cmd, "false")
	}
	rUri, err := url.Parse(r.Url)
	if err != nil {
		glog.Exitf("解析URL出错：%s", err.Error())
	}
	cmd = append(cmd, fmt.Sprintf("docker://%s/%s", rUri.Host, repository))
	return cmd
}
