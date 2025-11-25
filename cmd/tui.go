package cmd

import (
    "git-watcher/tui"
    "github.com/spf13/cobra"
)

var tuiPath string

var tuiCmd = &cobra.Command{
    Use:   "tui",
    Short: "Start terminal UI",
    RunE: func(cmd *cobra.Command, args []string) error {
        return tui.StartTUI(tuiPath)
    },
}

func init() {
    tuiCmd.Flags().StringVarP(&tuiPath, "path", "p", ".", "Directory path to scan")
    rootCmd.AddCommand(tuiCmd)
}
