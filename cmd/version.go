package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var version string = "development"

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "print version of gitlab-expiration-token",
	Long:  `print version of gitlab-expiration-token`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(version)
	},
}
