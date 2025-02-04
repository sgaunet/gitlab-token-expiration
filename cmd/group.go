package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/pterm/pterm"
	"github.com/sgaunet/gitlab-token-expiration/pkg/app"
	"github.com/sgaunet/gitlab-token-expiration/pkg/dto"
	"github.com/sgaunet/gitlab-token-expiration/pkg/gitlab"
	"github.com/sgaunet/gitlab-token-expiration/pkg/views"
	"github.com/spf13/cobra"
)

var noRecursiveOption bool

// groupCmd represents the group command to list expirable tokens of a group
var groupCmd = &cobra.Command{
	Use:   "group",
	Short: "List expirable tokens of a group",
	Long:  `List expirable tokens of a group`,
	Run: func(cmd *cobra.Command, args []string) {
		var tokens []dto.Token
		v := views.NewTableOutput(true, false)
		a := app.NewApp(v)

		// l := initTrace(os.Getenv("DEBUGLEVEL"))
		// a.SetLogger(l)
		ctx := context.Background()

		if gitlabID == 0 {
			fmt.Fprintln(os.Stderr, "You must provide a group ID")
			os.Exit(1)
		}

		if noRecursiveOption {
			// List only tokens of the group
			group, err := a.GetGroup(gitlabID)
			if err != nil {
				fmt.Fprintln(os.Stderr, err.Error())
				os.Exit(1)
			}
			tokens, err = a.GetTokensOfGroups(ctx, []gitlab.GitlabGroup{group})
			if err != nil {
				fmt.Fprintln(os.Stderr, err.Error())
				os.Exit(1)
			}
			v.Render(tokens)
		}

		if !noRecursiveOption {
			// List tokens of the group and its subgroups and projects
			// recursive option
			fmt.Println("Retrieve informations of all subgroups and projects")
			spinnerInfo, _ := pterm.DefaultSpinner.Start("Retrieve informations of all subgroups and projects")

			actualGroup, err := a.GetGroup(gitlabID)
			if err != nil {
				spinnerInfo.Fail("Error while retrieving group informations")
				fmt.Fprintln(os.Stderr, err.Error())
				os.Exit(1)
			}
			// List tokens of the group and its subgroups and projects
			groups, err := a.GetSubGroups(gitlabID)
			if err != nil {
				spinnerInfo.Fail("Error while retrieving subgroups")
				fmt.Fprintln(os.Stderr, err.Error())
				os.Exit(1)
			}
			groups = append(groups, actualGroup)

			projects, err := a.GetRecursiveProjectsOfGroup(gitlabID)
			if err != nil {
				spinnerInfo.Fail("Error while retrieving projects")
				fmt.Fprintln(os.Stderr, err.Error())
				os.Exit(1)
			}
			spinnerInfo.Success("Groups and projects retrieved")
			spinnerInfo, _ = pterm.DefaultSpinner.Start("Retrieve tokens of all subgroups and projects")

			tokens, err = a.GetTokensOfGroups(ctx, groups)
			if err != nil {
				spinnerInfo.Fail("Error while retrieving tokens of groups")
				fmt.Fprintln(os.Stderr, err.Error())
				os.Exit(1)
			}
			tokensOfProjects, err := a.GetTokensOfProjects(ctx, projects)
			if err != nil {
				spinnerInfo.Fail("Error while retrieving tokens of projects")
				fmt.Fprintln(os.Stderr, err.Error())
				os.Exit(1)
			}
			tokens = append(tokens, tokensOfProjects...)
			spinnerInfo.Success("Tokens retrieved")

			v.Render(tokens)
		}
	},
}
