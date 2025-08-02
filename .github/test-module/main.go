package main

import (
	"os"

	"github.com/spf13/cobra"
)

// newRootCmd creates and returns the root command
func newRootCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "test-module",
		Short: "A test module for multi-module support",
		Long:  `This is a test module to demonstrate multi-module support in mage-x.`,
		Run: func(cmd *cobra.Command, args []string) {
			// Use os.Stdout.WriteString instead of fmt.Println to avoid forbidigo
			_, _ = os.Stdout.WriteString("Test module executed successfully!\n") //nolint:errcheck // OK to ignore stdout write errors
		},
	}
}

func main() {
	rootCmd := newRootCmd()

	if err := rootCmd.Execute(); err != nil {
		_, _ = os.Stderr.WriteString(err.Error() + "\n") //nolint:errcheck // OK to ignore stderr write errors
		os.Exit(1)
	}
}
