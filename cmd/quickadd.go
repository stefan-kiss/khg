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
	"github.com/goccy/go-yaml"
	"github.com/spf13/viper"
	"github.com/stefan-kiss/khg/internal/cfg"
	"github.com/stefan-kiss/khg/internal/kubeconfig"
	"io/ioutil"
	"log"

	"github.com/spf13/cobra"
)

// quickaddCmd represents the quickadd command
var quickaddCmd = &cobra.Command{
	Use:   "quickadd",
	Short: "Adds a source to the config and gathers it also.",
	Long: `Adds a source to the config and gathers it also.
The options are the same as the ones for just adding a source
url examples        : ssh://10.0.0.1:2222/~/.kube.config
                      ssh://10.0.0.1:2222/./.kube.config
                      ~/projects/kubernetes/.kube.config

api-address examples: 10.0.0.1:10443
url path part must start with "/" so use "/./" for current directory and "~/" for home directory.
api-address must include the port.
`,
	Run: quickAdd,
}

func init() {
	rootCmd.AddCommand(quickaddCmd)

	// Here you will define your flags and configuration settings.

	quickaddCmd.Flags().StringP("label", "l", "", "Label for the entry. Will overwrite entry if exists.")
	quickaddCmd.Flags().StringP("url", "u", "", "Url for the config file. Can also be a filepath.")
	quickaddCmd.Flags().StringP("api-address", "a", "", "Use api address (usually external ip) instead of the one found in the source file.")
	quickaddCmd.Flags().BoolP("insecure", "i", false, "Will remove the CA from cluster and add the 'insecure-skip-tls-verify' flag.")
	quickaddCmd.Flags().BoolP("rewrite-api", "r", false, "Will rewrite api address using the host from the url and default 6443 port. Use api-address flag to overwrite this option and specify a custom one.")

	quickaddCmd.MarkFlagRequired("label")
	quickaddCmd.MarkFlagRequired("url")

}

func quickAdd(cmd *cobra.Command, args []string) {
	khg := cfg.Cfg{}
	err := viper.Unmarshal(&khg)
	if err != nil {
		log.Fatalf("unable to Unmarshal config file: %v", err)
	}

	label, err := cmd.Flags().GetString("label")
	if err != nil {
		log.Fatalf("unable get label from command line: %v", err)
	}
	src := cfg.Source{}
	url, err := cmd.Flags().GetString("url")
	if err != nil {
		log.Fatalf("unable get url from command line: %v", err)
	}
	src.Source = url

	insecure, err := cmd.Flags().GetBool("insecure")
	if err != nil {
		log.Fatalf("unable get insecure from command line: %v", err)
	}
	src.Insecure = insecure

	rewriteApi, err := cmd.Flags().GetBool("rewrite-api")
	if err != nil {
		log.Fatalf("unable get rewrite-api from command line: %v", err)
	}
	if rewriteApi {
		log.Fatal("rewrite-api flag not implemented yet")
	}

	apiAddress, err := cmd.Flags().GetString("api-address")
	if err != nil {
		log.Fatalf("unable get api-address from command line: %v", err)
	}
	if apiAddress != "" {
		src.ApiAddress = apiAddress
	}

	khg.Sources[label] = src

	configBytes, err := yaml.Marshal(khg)
	if err != nil {
		log.Fatalf("unable marshal the config file: %v", err)
	}

	err = ioutil.WriteFile(viper.ConfigFileUsed(), configBytes, 600)
	if err != nil {
		log.Fatalf("unable write the config file %s: %v", viper.ConfigFileUsed(), err)
	}

	destKonfig, err := kubeconfig.InitDestination(khg.Destination)
	if err != nil {
		log.Fatalf("unable to initialize destination file %s: %v", khg.Destination, err)
	}

	err = destKonfig.MergeOne(label, src)
	if err != nil {
		log.Fatalf("unable source into destination %s: %v", src.Source, err)
	}
}
