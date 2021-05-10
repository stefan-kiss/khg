/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stefan-kiss/khg/internal/cfg"
	"github.com/stefan-kiss/khg/internal/kubeconfig"
	"strings"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Lists the current contexts from the kubernetes config file.",
	Long: `Lists the current contexts from the kubernetes config file.
Optionally if the '-p/-persistent' flag is supplied the config file entries are also listed.

Matching with the config file label is done by extracting the label from the context name assuming the following format for it:
{{ initial_context_name }}@{{ label }}

`,
	Run: listCtx,
}

type listHead struct {
	ConfigLabel       string
	SourceUrl         string
	KubernetesContext string
	ApiAddress        string
}

func init() {
	rootCmd.AddCommand(listCmd)
}

func listCtx(cmd *cobra.Command, args []string) {
	table := []listHead{
		{
			ConfigLabel:       "ConfigLabel",
			SourceUrl:         "SourceUrl",
			KubernetesContext: "KubernetesContext",
			ApiAddress:        "ApiAddress",
		},
	}
	configUsed := cfg.Cfg{}
	err := viper.Unmarshal(&configUsed)
	if err != nil {
		log.Fatalf("unable to Unmarshal config file: %v", err)
	}

	destKonfig, err := kubeconfig.DestInit(configUsed.Destination)
	if err != nil {
		log.Fatalf("unable to initialize destination file %q: %v", configUsed.Destination, err)
	}

	for cfgLabel, source := range configUsed.Sources {
		foundCtx := ""
		for ctxLabel, context := range destKonfig.Config.Contexts {
			if strings.HasSuffix(ctxLabel, "@"+cfgLabel) {
				foundCtx = ctxLabel
				table = append(table, listHead{
					ConfigLabel:       cfgLabel,
					SourceUrl:         source.Source,
					KubernetesContext: ctxLabel,
					ApiAddress:        destKonfig.Config.Clusters[context.Cluster].Server,
				})
			}
			break
		}
		if foundCtx != "" {
			delete(destKonfig.Config.Contexts, foundCtx)
		} else {
			table = append(table, listHead{
				ConfigLabel: cfgLabel,
				SourceUrl:   source.Source,
			})
		}
	}
	for ctxLabel, context := range destKonfig.Config.Contexts {
		table = append(table, listHead{
			ConfigLabel:       "",
			SourceUrl:         "",
			KubernetesContext: ctxLabel,
			ApiAddress:        destKonfig.Config.Clusters[context.Cluster].Server,
		})
	}
	for _, tblElem := range table {
		fmt.Printf("%-20s | %-40s | %-50s | %-20s\n",
			tblElem.ConfigLabel,
			tblElem.SourceUrl,
			tblElem.KubernetesContext,
			tblElem.ApiAddress,
		)
	}
}
