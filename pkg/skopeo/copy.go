package skopeo

import (
	"fmt"
	"github.com/TimeBye/registry-manager/pkg/global"
	"github.com/go-cmd/cmd"
	"github.com/x-mod/glog"
	"net/url"
	"strings"
	"sync"
	"time"
)

func Copy(repository, tag string, wg *sync.WaitGroup) {
	defer wg.Done()
	retryStart := 1
	cmdArgs := generateCopyArgs(repository, tag)
RePlay:
	if retryStart > global.Retry {
		global.FailedList[cmdArgs[len(cmdArgs)-2]] = cmdArgs[len(cmdArgs)-1]
		glog.Errorf("同步镜像已达最大重试次数：%d，%s to %s",
			global.Retry, cmdArgs[len(cmdArgs)-2], cmdArgs[len(cmdArgs)-1])
	} else {
		glog.Infof("当前同步：%s to %s", cmdArgs[len(cmdArgs)-2], cmdArgs[len(cmdArgs)-1])
		glog.V(5).Infof("skopeo %s", strings.Join(cmdArgs, " "))
		skopeoCmd := cmd.NewCmd("skopeo", cmdArgs...)
		skopeoCmd.Start()
		ticker := time.NewTicker(2 * time.Second)
		n := 0
		for range ticker.C {
			status := skopeoCmd.Status()
			if !status.Complete {
				if n = len(status.Stdout); n > 0 {
					glog.Info(status.Stdout[n-1])
				}
			} else {
				if status.Exit != 0 {
					glog.Errorf("同步镜像出错：%s", strings.Join(status.Stderr, ""))
					retryStart = retryStart + 1
					goto RePlay
				}
				break
			}
		}
	}
}

func generateCopyArgs(repository, tag string) []string {
	fromR := global.Manager.Registries[global.Manager.SyncPolicy.From]
	toR := global.Manager.Registries[global.Manager.SyncPolicy.To]

	cmd := make([]string, 0)
	cmd = append(cmd, "copy")
	cmd = append(cmd, "--all")
	if fromR.Username != "" && fromR.Password != "" {
		cmd = append(cmd, "--src-creds")
		cmd = append(cmd, fmt.Sprintf("%s:%s", fromR.Username, fromR.Password))
	}
	if toR.Username != "" && toR.Password != "" {
		cmd = append(cmd, "--dest-creds")
		cmd = append(cmd, fmt.Sprintf("%s:%s", toR.Username, toR.Password))
	}
	if fromR.Insecure {
		cmd = append(cmd, "--src-tls-verify")
		cmd = append(cmd, "false")
	}
	if toR.Insecure {
		cmd = append(cmd, "--dest-tls-verify")
		cmd = append(cmd, "false")
	}
	fUri, err := url.Parse(fromR.Url)
	if err != nil {
		glog.Exitf("解析URL出错：%s", err.Error())
	}
	tUri, err := url.Parse(toR.Url)
	if err != nil {
		glog.Exitf("解析URL出错：%s", err.Error())
	}
	cmd = append(cmd, fmt.Sprintf("docker://%s/%s:%s", fUri.Host, repository, tag))
	cmd = append(cmd, fmt.Sprintf("docker://%s/%s:%s", tUri.Host,
		global.Manager.SyncPolicy.ReplaceName(repository), tag))
	return cmd
}
