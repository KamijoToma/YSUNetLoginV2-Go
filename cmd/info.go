package cmd

import (
	"fmt"

	"ruijie-go/internal/client"
	"ruijie-go/internal/config"
	"ruijie-go/internal/utils"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// infoCmd represents the info command
var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Show account information",
	Long:  `Show detailed account information for the current logged-in user.`,
	RunE:  runInfo,
}

func init() {
	rootCmd.AddCommand(infoCmd)
}

func runInfo(cmd *cobra.Command, args []string) error {
	// Create configuration
	cfg := config.NewConfig()
	cfg.LoadFromViper()
	cfg.UpdateFromFlags("", "", "", viper.GetString("proxy"), viper.GetBool("verbose"))

	// Create Ruijie client
	ruijieClient := client.NewRuijieClient(cfg.Proxies, cfg.Verbose)

	// First check if logged in
	isLoggedIn, userInfo, err := ruijieClient.CheckLoginStatus()
	if err != nil {
		fmt.Printf("Error: %s\n", config.GetErrorMessage(err))
		return err
	}

	if !isLoggedIn {
		fmt.Println("Error: Not logged in. Please login first.")
		return fmt.Errorf("not logged in")
	}

	// Get session information
	sessionInfo, err := ruijieClient.RedirectToPortal("")
	if err != nil {
		fmt.Printf("Error: %s\n", config.GetErrorMessage(err))
		return err
	}

	// Get account information
	accountInfo, err := ruijieClient.GetAccountInfo(sessionInfo)
	if err != nil {
		fmt.Printf("Error: %s\n", config.GetErrorMessage(err))
		return err
	}

	// Print user status information
	if userInfoMap, ok := userInfo.(map[string]interface{}); ok {
		utils.PrintStatusInfo(userInfoMap)
	}
	fmt.Println()

	// Print account information
	utils.PrintAccountInfo(accountInfo)

	return nil
}
