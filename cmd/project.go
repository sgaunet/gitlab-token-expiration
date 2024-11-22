package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/sgaunet/gitlab-token-expiration/pkg/app"
	"github.com/sgaunet/gitlab-token-expiration/pkg/views"
	"github.com/spf13/cobra"
)

// projectCmd represents the project command to list expirable tokens of a project
var projectCmd = &cobra.Command{
	Use:   "project",
	Short: "List expirable tokens of a project",
	Long:  `List expirable tokens of a project`,
	Run: func(cmd *cobra.Command, args []string) {
		v := views.NewTableOutput(true, false)
		a := app.NewApp(v)

		// l := initTrace(os.Getenv("DEBUGLEVEL"))
		// a.SetLogger(l)
		ctx := context.Background()

		if gitlabID == 0 {
			fmt.Fprintln(os.Stderr, "You must provide a group ID")
			os.Exit(1)
		}

		project, err := a.GetProject(gitlabID)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
		err = a.ListTokensOfProject(ctx, project)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
	},
}
