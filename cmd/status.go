package cmd

import (
	"fmt"

	"ruijie-go/internal/client"
	"ruijie-go/internal/config"
	"ruijie-go/internal/utils"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check login status",
	Long:  `Check the current login status of the Ruijie network authentication system.`,
	RunE:  runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func runStatus(cmd *cobra.Command, args []string) error {
	// Create configuration
	cfg := config.NewConfig()
	cfg.LoadFromViper()
	cfg.UpdateFromFlags("", "", "", viper.GetString("proxy"), viper.GetBool("verbose"))

	// Create Ruijie client
	ruijieClient := client.NewRuijieClient(cfg.Proxies, cfg.Verbose)

	// Check login status
	isLoggedIn, info, err := ruijieClient.CheckLoginStatus()
	if err != nil {
		fmt.Printf("Error: %s\n", config.GetErrorMessage(err))
		return err
	}

	if isLoggedIn {
		if userInfo, ok := info.(map[string]interface{}); ok {
			utils.PrintStatusInfo(userInfo)
		} else {
			fmt.Println("Online (status information unavailable)")
		}
	} else {
		fmt.Println("Offline")
	}

	return nil
}
