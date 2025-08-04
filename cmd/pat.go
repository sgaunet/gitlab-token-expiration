package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/sgaunet/gitlab-token-expiration/pkg/app"
	"github.com/sgaunet/gitlab-token-expiration/pkg/views"
	"github.com/spf13/cobra"
)

// patCmd represents the command to list personal access tokens.
var patCmd = &cobra.Command{
	Use:   "pat",
	Short: "List gitlab personal access tokens",
	Long:  `List personal access tokens from gitlab`,
	Run: func(_ *cobra.Command, _ []string) {
		v := views.NewTableOutput(views.WithColorOption(!printNoColor),
			views.WithHeaderOption(!printNoHeader),
			views.WithPrintRevokedOption(printRevoked),
			views.WithNbDaysBeforeExp(nbDaysBeforeExp),
		)
		a := app.NewApp(v, app.WithRevokedToken(printRevoked))

		// l := initTrace(os.Getenv("DEBUGLEVEL"))
		// a.SetLogger(l)
		ctx := context.Background()

		tokens, err := a.GetPersonalAccessTokens(ctx)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
		if err := v.Render(tokens); err != nil {
			fmt.Fprintf(os.Stderr, "Error rendering tokens: %v\n", err)
			os.Exit(1)
		}
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
