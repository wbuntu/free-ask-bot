package cmd

import (
	"fmt"

	"gitbub.com/wbuntu/free-ask-bot/internal/pkg/config"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the free-ask-bot version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(config.C.Version)
	},
}
