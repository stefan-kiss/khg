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
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stefan-kiss/khg/internal/cfg"
	"github.com/stefan-kiss/khg/internal/kubeconfig"
)

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:   "delete <context>",
	Args:  cobra.ExactArgs(1),
	Short: "Deletes the local kubernetes configuration for the supplied context name.",
	Long: `Deletes the local kubernetes configuration for the supplied context name.
Optionally if the '-p/-persistent' flag is supplied and a matching configuration can be found in the config file it is deleted also.

Matching with the config file label is done by extracting the label from the context name assuming the following format for it:
{{ initial_context_name }}@{{ label }}
`,
	Run: deleteCtx,
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}

func deleteCtx(cmd *cobra.Command, args []string) {
	label := args[0]
	log.Infof("deleteing label: %q", label)

	configUsed := cfg.Cfg{}
	err := viper.Unmarshal(&configUsed)
	if err != nil {
		log.Fatalf("unable to Unmarshal config file: %v", err)
	}
	persistent, err := cmd.Flags().GetBool("persistent")
	if err != nil {
		log.Fatalf("unable to get 'persistent' flag value")
	}

	if persistent {
		log.Infof("deleting label: %q from persistent config file", label)
		err = cfg.Delete(&configUsed, label)
		if err != nil {
			log.Errorf("unable to delete label: %q from persistent config: %v. continuing", label, err)
		}
		err = cfg.Save(&configUsed)
		if err != nil {
			log.Fatalf("unable to save persistent config: %s, %v", viper.ConfigFileUsed(), err)
		}
	}

	destKonfig, err := kubeconfig.DestInit(configUsed.Destination)
	if err != nil {
		log.Fatalf("unable to initialize destination file %q: %v", configUsed.Destination, err)
	}
	log.Infof("deleting label: %q from kubernetes config file", label)
	err = destKonfig.Delete(label)
	if err != nil {
		log.Fatalf("unable to delete label: %q from: %q: %v", label, destKonfig.Url, err)
	}
	log.Infof("succesfuly deleted: %q", label)

}
