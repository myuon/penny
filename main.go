package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var version = "0.1.0"

func main() {
	rootCmd := &cobra.Command{
		Use:     "penny",
		Short:   "penny - a CLI tool",
		Long:    `penny is a command line tool for managing your tasks.`,
		Version: version,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
