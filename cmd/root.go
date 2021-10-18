// Copyright (c) 2021. Stefan Kiss
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
//
package cmd

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/stefan-kiss/khg/internal/cfg"
	"os"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "khg",
	Short: "Command line app to gather (and unify) kubernetes config files from various sources.",
	Long: `Kubernetes hunter-gather is a command line application that serves several purposes:

Fetch kubernetes configuration files from various sources
Transform them according to configurable rules (such as change labels,api endpoints ... etc)
Merge them in one config file ready to be used by tools like kubectx


`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(cmd *cobra.Command, args []string) { },
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
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.khg.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.PersistentFlags().BoolP("persistent", "p", false, "persist any changes to config file")
	rootCmd.PersistentFlags().StringP("identity", "I", "", "ssh private key")
	rootCmd.PersistentFlags().StringP("log-level", "L", "INFO", "Log Level. Default INFO")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	levelStr, err := rootCmd.PersistentFlags().GetString("log-level")
	if err != nil {
		log.Fatalf("unable to get log-level: %q: %q", levelStr, err)
	}

	logLevel, err := log.ParseLevel(levelStr)
	if err != nil {
		log.Fatalf("unable to parse log-level: %q: %q", levelStr, err)
	}
	log.SetLevel(logLevel)

	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".khg" (without extension).
		viper.AddConfigPath(home)

		// Search current dir ".khg" (without extension).
		currentDir, err := os.Getwd()
		if err == nil {
			viper.AddConfigPath(currentDir)
		}
		viper.SetConfigName(".khg")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
	khg := cfg.Cfg{}
	viper.Unmarshal(&khg)
}
