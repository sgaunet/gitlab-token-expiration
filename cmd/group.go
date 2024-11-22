package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/sgaunet/gitlab-token-expiration/pkg/app"
	"github.com/sgaunet/gitlab-token-expiration/pkg/views"
	"github.com/spf13/cobra"
)

var recursiveOption bool

// groupCmd represents the group command to list expirable tokens of a group
var groupCmd = &cobra.Command{
	Use:   "group",
	Short: "List expirable tokens of a group",
	Long:  `List expirable tokens of a group`,
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

		group, err := a.GetGroup(gitlabID)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
		err = a.ListTokensOfGroup(ctx, group)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
	},
}
