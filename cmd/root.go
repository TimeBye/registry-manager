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
	"flag"
	"fmt"
	"github.com/spf13/pflag"
	"os"

	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/TimeBye/registry-manager/pkg/types"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "registry-manager",
	Short: "Docker 镜像库管理工具",
	Long:  `registry-manager 是一个 Docker 镜像库管理小工具，可以实现按指定规则删除镜像 tag`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) {},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	flag.CommandLine.Parse([]string{})
	flag.Set("logtostderr", "true")

	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	rootCmd.PersistentFlags().StringVarP(&types.CfgFile, "config", "c", "",
		"配置文件路径，默认当前目录下 registry-manager.yml 文件")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if types.CfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(types.CfgFile)
	} else {
		// Search config in home directory with name ".registry-manager" (without extension).
		viper.AddConfigPath(".")
		viper.SetConfigName("registry-manager")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		glog.Infof("使用配置文件: %s", viper.ConfigFileUsed())
	}
	err := viper.Unmarshal(&types.Manager)
	if err != nil {
		panic(err)
	}
	types.Manager.Init()
}
