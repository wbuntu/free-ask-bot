package cmd

import (
	"fmt"

	"gitbub.com/wbuntu/gin-template/internal/pkg/config"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the gin-template version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(config.C.Version)
	},
}
