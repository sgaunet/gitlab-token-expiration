package cmd

import "github.com/spf13/pflag"

// RootCmd exposes the root command to the black-box tests in cmd_test.
var RootCmd = rootCmd

// ResetFlags restores every flag of every command to its default value. The
// flag variables are package globals, so a test run must not leak into the next.
func ResetFlags() {
	for _, c := range rootCmd.Commands() {
		c.Flags().VisitAll(func(f *pflag.Flag) {
			_ = f.Value.Set(f.DefValue)
			f.Changed = false
		})
	}
}
