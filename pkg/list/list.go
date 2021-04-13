package list

import (
	"fmt"
	client "github.com/TimeBye/docker-registry-client/registry"
	"github.com/TimeBye/registry-manager/pkg/global"
	"github.com/TimeBye/registry-manager/pkg/skopeo"
	"github.com/x-mod/glog"
)

func Run() {
	registries := global.Manager.Registries
	for _, registry := range registries {
		registryClient := &client.Registry{}
		if !registry.Insecure {
			registryClient, _ = client.New(registry.Url, registry.Username, registry.Password)
		} else {
			registryClient, _ = client.NewInsecure(registry.Url, registry.Username, registry.Password)
		}
		repositories, err := registryClient.Repositories()
		if err != nil {
			glog.Exitf("获取仓库出错：%s", err.Error())
		}
		for _, repository := range repositories {
			tags := skopeo.Tags(registry, repository)
			for _, tag := range tags.Tags {
				fmt.Println(fmt.Sprintf("%s:%s", tags.Repository, tag))
			}
		}
	}
}
