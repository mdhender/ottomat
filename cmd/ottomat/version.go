package main

import (
	"fmt"

	"github.com/mdhender/ottomat"
	"github.com/spf13/cobra"
)

var (
	buildInfo bool = false
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Long:  `Display the current version of OttoMat.`,
	Run: func(cmd *cobra.Command, args []string) {
		if buildInfo {
			fmt.Println(ottomat.Version().String())
		} else {
			fmt.Println(ottomat.Version().Core())
		}
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
	versionCmd.Flags().BoolVar(&buildInfo, "build-info", false, "show build information")
}
