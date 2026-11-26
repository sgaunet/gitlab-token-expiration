package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// patCmd represents the command to list personal access tokens.
var patCmd = &cobra.Command{
	Use:   "pat",
	Short: "List your gitlab personal access tokens",
	Long: `List your personal access tokens.

Only your own tokens are listed, even with an administrator token.`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		cmd.SilenceUsage = true
		out, err := prepare()
		if err != nil {
			return err
		}
		a, err := newApp()
		if err != nil {
			return err
		}
		tokens, err := a.GetPersonalAccessTokens(cmd.Context())
		if err != nil {
			return fmt.Errorf("retrieving personal access tokens: %w", err)
		}
		return out.render(a, tokens)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
	// Here you will define your flags and configuration settings.
	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// versionCmd.PersistentFlags().String("foo", "", "A help for foo")
	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// versionCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
