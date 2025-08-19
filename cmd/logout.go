package cmd

import (
	"fmt"

	"ruijie-go/internal/client"
	"ruijie-go/internal/config"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// logoutCmd represents the logout command
var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Logout from network",
	Long:  `Logout from the Ruijie network authentication system.`,
	RunE:  runLogout,
}

func init() {
	rootCmd.AddCommand(logoutCmd)
}

func runLogout(cmd *cobra.Command, args []string) error {
	// Create configuration
	cfg := config.NewConfig()
	cfg.LoadFromViper()
	cfg.UpdateFromFlags("", "", "", viper.GetString("proxy"), viper.GetBool("verbose"))

	// Create Ruijie client
	ruijieClient := client.NewRuijieClient(cfg.Proxies, cfg.Verbose)

	// Execute logout
	if err := ruijieClient.Logout(); err != nil {
		fmt.Printf("Error: %s\n", config.GetErrorMessage(err))
		return err
	}

	fmt.Println("Logout successful.")
	return nil
}
