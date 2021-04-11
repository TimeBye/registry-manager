package list

import (
	"fmt"
	"github.com/TimeBye/docker-registry-client/registry"
	"github.com/TimeBye/registry-manager/pkg/global"
	"github.com/TimeBye/registry-manager/pkg/skopeo"
	"github.com/x-mod/glog"
)

func Run() {
	registries := global.Manager.Registries
	for k, v := range registries {
		registryClient := &registry.Registry{}
		if !v.Insecure {
			registryClient, _ = registry.New(v.Url, v.Username, v.Password)
		} else {
			registryClient, _ = registry.NewInsecure(v.Url, v.Username, v.Password)
		}
		var err error
		repositories, err := registryClient.Repositories()
		if err != nil {
			glog.Error(err)
		}
		for _, repo := range repositories {
			tags := skopeo.Tags(k, repo)
			for _, tag := range tags.Tags {
				fmt.Println(fmt.Sprintf("%s:%s",
					tags.Repository, tag))
			}
		}
	}
}
