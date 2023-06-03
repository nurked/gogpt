package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var modelCmd = &cobra.Command{
	Use:   "model",
	Short: "Change the model",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			fmt.Println("Please provide a model")
			os.Exit(1)
		}

		viper.Set("MODEL", args[0])
		err := viper.WriteConfig()
		if err != nil {
			fmt.Printf("Error saving config: %s", err)
			os.Exit(1)
		}
	},
}
