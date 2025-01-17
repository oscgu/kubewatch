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
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var (
	buildDate, gitCommit string
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "print version",
	Long:  `print version`,
	Run: func(cmd *cobra.Command, args []string) {
		versionPrettyString()
	},
}

func versionPrettyString() {
	log.Info().Msgf("gitCommit: %s", gitCommit)
	log.Info().Msgf("buildDate: %s", buildDate)
}

func init() {
	RootCmd.AddCommand(versionCmd)
}
