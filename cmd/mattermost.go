/*
Copyright 2016 Skippbox, Ltd.

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
	"github.com/bitnami-labs/kubewatch/config"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// mattermostConfigCmd represents the mattermost subcommand
var mattermostConfigCmd = &cobra.Command{
	Use:   "mattermost",
	Short: "specific mattermost configuration",
	Long:  `specific mattermost configuration`,
	Run: func(cmd *cobra.Command, args []string) {
		conf, err := config.New()
		if err != nil {
			log.Fatal().Err(err).Send()
		}

		channel, err := cmd.Flags().GetString("channel")
		if err == nil {
			if len(channel) > 0 {
				conf.Handler.Mattermost.Channel = channel
			}
		} else {
			log.Fatal().Err(err).Send()
		}

		url, err := cmd.Flags().GetString("url")
		if err == nil {
			if len(url) > 0 {
				conf.Handler.Mattermost.Url = url
			}
		} else {
			log.Fatal().Err(err).Send()
		}

		username, err := cmd.Flags().GetString("username")
		if err == nil {
			if len(url) > 0 {
				conf.Handler.Mattermost.Username = username
			}
		} else {
			log.Fatal().Err(err).Send()
		}

		if err = conf.Write(); err != nil {
			log.Fatal().Err(err).Send()
		}
	},
}

func init() {
	mattermostConfigCmd.Flags().StringP("channel", "c", "", "Specify Mattermost channel")
	mattermostConfigCmd.Flags().StringP("url", "u", "", "Specify Mattermost url")
	mattermostConfigCmd.Flags().StringP("username", "n", "", "Specify Mattermost username")
}
