package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var version = "development"

// versionCmd represents the version command.
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "print version of gitlab-expiration-token",
	Long:  `print version of gitlab-expiration-token`,
	Run: func(_ *cobra.Command, _ []string) {
		fmt.Println(version)
	},
}
