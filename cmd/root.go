// Copyright © 2019 NAME HERE <EMAIL ADDRESS>
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

package cmd

import (
	"bytes"
	"github.com/TimeBye/registry-manager/pkg/global"
	"github.com/TimeBye/registry-manager/pkg/types"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/x-mod/glog"
	"gopkg.in/yaml.v2"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "registry-manager",
	Short: "Docker 镜像库管理工具",
	Long:  `registry-manager 是一个 Docker 镜像库管理小工具`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) {},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		glog.Exitf("执行出错：%s", err.Error())
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().IntVarP(&global.Verbosity, "verbosity", "v", 1, "INFO信息显示级别")
	rootCmd.PersistentFlags().IntVarP(&global.Retry, "retry", "r", 3, "失败重试次数")
	rootCmd.PersistentFlags().IntVarP(&global.ProcessLimit, "process-limit", "p", 3, "并发同步数量")
	rootCmd.PersistentFlags().StringVarP(
		&global.CfgFile, "config", "c", "", "config file (default is ./config.yaml)")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	glog.Open(glog.LogToStderr(true),
		glog.Verbosity(global.Verbosity),
	)
	defer glog.Close()
	if global.CfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(global.CfgFile)
	} else {
		// Search config in home directory with name ".sync" (without extension).
		viper.AddConfigPath(".")
		viper.SetConfigName("config")
	}
	global.Manager = &types.Config{
		DeletePolicy: types.DeletePolicy{
			Start:    0,
			DryRun:   true,
			MixCount: 10,
		},
		SyncPolicy: types.SyncPolicy{
			DryRun: true,
			Filters: []string{".*"},
		},
	}
	b, err := yaml.Marshal(global.Manager)
	if err != nil {
		panic(err)
	}
	defaultConfig := bytes.NewReader(b)
	if err := viper.MergeConfig(defaultConfig); err != nil {
		panic(err)
	}

	viper.AutomaticEnv() // read in environment variables that match
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := viper.MergeInConfig(); err != nil {
		glog.Exitf("合并yaml出错：%s", err.Error())
	}

	err = viper.Unmarshal(global.Manager)
	if err != nil {
		glog.Exitf("解析yaml出错：%s", err.Error())
	}

	for _, registry := range global.Manager.Registries {
		registry.Uri, err = url.Parse(registry.Url)
		if err != nil {
			glog.Exitf("解析URL出错：%s", err.Error())
		}
	}

	glog.V(4).Infof("%+v", global.Manager)
}
