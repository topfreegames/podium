// santiago - webhook dispatching service
// https://github.com/topfreegames/santiago
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/topfreegames/podium/api"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "returns Podium version",
	Long:  `returns Podium version`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Podium v%s\n", api.VERSION)
	},
}

func init() {
	RootCmd.AddCommand(versionCmd)
}
