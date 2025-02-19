/*
Copyright © 2020 Google Inc.

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
	"gkekitctl/pkg/acm"
	"gkekitctl/pkg/analytics"
	"gkekitctl/pkg/config"
	"gkekitctl/pkg/lifecycle"

	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create GKE Demo Environment",
	Example: ` gkekitctl create
	gkekitctl create --config <file.yaml>`,
	Run: func(cmd *cobra.Command, args []string) {
		conf := config.InitConf(cfgFile)

		// Send user analytics - async
		if conf.SendAnalytics {
			go analytics.SendAnalytics(conf, Version, GitCommit)
		}

		config.GenerateTfvars(conf)
		tfStateBucket, err := config.CheckTfStateType(conf)
		if err != nil {
			log.Fatalf("Error checking Tf State type: %s", err)
		}

		if conf.VpcConfig.VpcType == "shared" {
			lifecycle.InitTF("shared_vpc", tfStateBucket[1], conf.VpcConfig.VpcType)
			lifecycle.ApplyTF("shared_vpc")
		}
		lifecycle.InitTF("cluster_build", tfStateBucket[0], conf.VpcConfig.VpcType)
		lifecycle.ApplyTF("cluster_build")

		// Init ACM (either ConfigSync or ConfigSync plus PolicyController)
		if conf.ConfigSync {
			err := acm.InitACM(conf)
			if err != nil {
				log.Errorf("🚨 Failed to initialize ACM: %s", err)
			}
		}

	},
}

func init() {
	rootCmd.AddCommand(createCmd)
}
