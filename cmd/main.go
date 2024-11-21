package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/sgaunet/gitlab-token-expiration/pkg/app"
	"github.com/sgaunet/gitlab-token-expiration/pkg/views"
)

var version string = "development"

func printVersion() {
	fmt.Println(version)
}

func main() {
	var gitlabProject, gitlabGroup int
	var vOption bool
	var helpOption bool
	var err error

	// Parameters treatment
	flag.IntVar(&gitlabProject, "p", 0, "Gitlab project ID")
	flag.IntVar(&gitlabGroup, "g", 0, "Gitlab group ID")
	flag.BoolVar(&vOption, "v", false, "Get version")
	flag.BoolVar(&helpOption, "h", false, "help")
	flag.Parse()

	if helpOption {
		flag.Usage()
		os.Exit(0)
	}

	if vOption {
		printVersion()
		os.Exit(0)
	}

	if gitlabGroup != 0 && gitlabProject != 0 {
		fmt.Println("You must provide a Gitlab project ID or a Gitlab group ID, not both")
		flag.Usage()
		os.Exit(1)
	}

	gitlabToken := os.Getenv("GITLAB_TOKEN")
	if gitlabToken == "" {
		fmt.Println("You must provide a Gitlab token")
		flag.Usage()
		os.Exit(1)
	}

	// v := views.NewTerminalOutput(true, false)
	v := views.NewTableOutput(true, false)
	a := app.NewApp(v)

	l := initTrace(os.Getenv("DEBUGLEVEL"))
	a.SetLogger(l)
	ctx := context.Background()

	if gitlabGroup == 0 && gitlabProject == 0 {
		err = a.ListPersonalAccessTokens(ctx)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
		os.Exit(0)
	}

	if gitlabProject != 0 {
		project, err := a.GetProject(gitlabProject)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
		err = a.ListTokensOfProject(ctx, project)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
	}

	if gitlabGroup != 0 {
		group, err := a.GetGroup(gitlabGroup)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
		err = a.ListTokensOfGroup(ctx, group)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
	}

}
