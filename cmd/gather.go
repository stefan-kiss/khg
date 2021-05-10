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
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stefan-kiss/khg/internal/cfg"
	"github.com/stefan-kiss/khg/internal/kubeconfig"
	"log"
)

// gatherCmd represents the gather command
var gatherCmd = &cobra.Command{
	Use:   "gather",
	Short: "Automaticaly discover and merge all defined configs",
	Long: `Gather is the main action and probably the one you will use most of the times.
It reads the configuration file and then reads, modifies and merge each kubeconfig into the destination.`,
	Run: gather,
}

func init() {
	rootCmd.AddCommand(gatherCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// gatherCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// gatherCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func gather(cmd *cobra.Command, args []string) {
	var err error

	khg := cfg.Cfg{}
	viper.Unmarshal(&khg)

	dest, err := kubeconfig.DestInit(khg.Destination)
	if err != nil {
		log.Fatalf("unable to parse destination url: %v: %v", khg.Destination, err)
	}

	err = dest.ReadConfig()
	if err != nil {
		log.Fatalf("unable to parse destination config file: %v: %v", khg.Destination, err)
	}

	konfigs := make([]*kubeconfig.KubeConfig, 0)
	for label, src := range khg.Sources {
		k, err := kubeconfig.DestInit(src.Source)
		if err != nil {
			log.Fatalf("unable to parse destination: %v: %v", khg.Destination, err)
		}
		k.Label = label
		k.SrcDef = src
		konfigs = append(konfigs, k)
	}

	for _, konfig := range konfigs {
		err = dest.CopyCurrentContext(konfig)
		if err != nil {
			log.Fatalf("unable merge config: %v: %v", konfig.Url, err)
		}

		err = dest.WriteConfig()
		if err != nil {
			log.Fatalf("unable write config: %v: %v", dest.Url, err)
		}

	}
}
