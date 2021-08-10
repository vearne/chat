// Copyright © 2018 NAME HERE <EMAIL ADDRESS>
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
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/vearne/chat/config"
	"github.com/vearne/chat/consts"
	"log"
	"os"
)

var (
	// config file path
	cfgFile     string
	versionFlag bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use: "chat <command>",
	Run: func(cmd *cobra.Command, args []string) {
		if versionFlag {
			fmt.Println("service: chat")
			fmt.Println("Version", consts.Version)
			fmt.Println("BuildTime", consts.BuildTime)
			fmt.Println("GitTag", consts.GitTag)
		}
	},
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
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	flagset := rootCmd.PersistentFlags()
	flagset.StringVar(&cfgFile, "config", "", "config file")

	// build info
	flagset.BoolVarP(&versionFlag, "version", "v", false, "Show version")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig(role string) {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)

	} else {
		viper.AddConfigPath("config")
		fname := fmt.Sprintf("config.%s", role)
		viper.SetConfigName(fname)
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		log.Println("Using config file:", viper.ConfigFileUsed())
	} else {
		log.Println("can't find config file", err)
	}
	config.InitConfig()
}
