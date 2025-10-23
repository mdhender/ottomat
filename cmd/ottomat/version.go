package main

import (
	"fmt"

	"github.com/maloquacious/semver"
	"github.com/spf13/cobra"
)

var (
	buildInfo bool = false
	version        = semver.Version{Major: 0, Minor: 2, Patch: 4, PreRelease: "alpha", Build: semver.Commit()}
)

func Version() semver.Version {
	return version
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Long:  `Display the current version of OttoMat.`,
	Run: func(cmd *cobra.Command, args []string) {
		if buildInfo {
			fmt.Println(Version().String())
		} else {
			fmt.Println(Version().Core())
		}
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
	versionCmd.Flags().BoolVar(&buildInfo, "build-info", false, "show build information")
}
