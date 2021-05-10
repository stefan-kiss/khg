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
	"github.com/k0kubun/pp"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/stefan-kiss/khg/internal/cfg"
	"github.com/stefan-kiss/khg/internal/kubeconfig"

	"github.com/spf13/cobra"
)

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get <source>",
	Short: "Gets kube configuration from a source.",
	Long: `Gets kube configuration from a source. Updating the configuration file is controlled by the 'persistent' global flag.
The options are the same as the ones for just adding a source
url examples        : 10.0.0.1:2222/~/.kube.config
                      ssh://centos@10.0.0.1:2222/./.kube.config
                      ssh://centos@10.0.0.1
					  10.0.0.1
                      example.com
                      file://~/projects/kubernetes/.kube.config

api-address examples: 10.0.0.1:10443
If using ssh protocol the url path part must start with "/" so use "/./" for current directory and "/~/" for home directory.
api-address must include the port.
Note: autodetecting api-address for file:// sources is currently broken and will be addressed in a later version.
`,
	Args: cobra.ExactArgs(1),
	Run:  add,
}

func init() {
	rootCmd.AddCommand(getCmd)

	// Here you will define your flags and configuration settings.

	getCmd.Flags().StringP("label", "l", "", "Label for the entry. Will overwrite entry if exists.")
	getCmd.Flags().StringP("api-address", "a", "", "Use api address (usually external ip) instead of the one found in the source file.")
	getCmd.Flags().BoolP("insecure", "i", false, "Will remove the CA from cluster and add the 'insecure-skip-tls-verify' flag.")
	getCmd.Flags().BoolP("rewrite-api", "r", false, "Will rewrite api address using the host from the url and default port. Use api-address flag to overwrite this option and specify a custom one.")

}

func add(cmd *cobra.Command, args []string) {
	configUsed := cfg.Cfg{}
	err := viper.Unmarshal(&configUsed)
	if err != nil {
		log.Fatalf("unable to Unmarshal config file: %v", err)
	}

	src := cfg.Source{}
	src.Source = args[0]

	label, err := cmd.Flags().GetString("label")
	if err != nil {
		log.Fatalf("unable get label from command line: %v", err)
	}

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
		log.Warn("rewrite-api flag is set. we will try to autodetect and rewrite api address. insecure is implied.")
		src.Insecure = true
		src.AutodetectApi = true
	}

	apiAddress, err := cmd.Flags().GetString("api-address")
	if err != nil {
		log.Fatalf("unable get api-address from command line: %v", err)
	}
	if apiAddress != "" {
		if rewriteApi {
			log.Fatal("rewrite-api will try to autodetect api address. remove it if you want to specify it yourself")
		}
		src.ApiAddress = apiAddress
	}

	sourceKonfig, err := kubeconfig.SourceInit(src, label)
	if err != nil {
		log.Fatalf("unable to parse source: %v: %v", src.Source, err)
	}
	if label != "" {
		sourceKonfig.Label = label
	} else {
		sourceKonfig.Label = sourceKonfig.Url.Host
	}
	log.Debugf("label: %s", sourceKonfig.Label)

	destKonfig, err := kubeconfig.DestInit(configUsed.Destination)
	if err != nil {
		log.Fatalf("unable to initialize destination file %s: %v", configUsed.Destination, err)
	}

	err = destKonfig.MergeOne(sourceKonfig)
	if err != nil {
		log.Fatalf("unable source into destination %s: %v", src.Source, err)
	}

	persistent, err := rootCmd.Flags().GetBool("persistent")
	if err != nil {
		log.Fatalf("unable get persistent flag: %v", err)
	}

	pp.Println(sourceKonfig.SrcDef)
	if persistent {
		err := cfg.Add(&configUsed, sourceKonfig.Label, sourceKonfig.SrcDef)
		if err != nil {
			log.Fatalf("unable to save config file: %v", err)
		}
	}
}
