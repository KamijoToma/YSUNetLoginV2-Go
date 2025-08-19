package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	verbose bool
	proxy   string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "ruijie-go",
	Short: "燕山大学锐捷V2网络认证命令行工具",
	Long: `燕山大学锐捷V2网络认证命令行工具

Examples:
  ruijie-go login -u 1145141919810 -p mypassword
  ruijie-go login  # Interactive login
  ruijie-go status
  ruijie-go logout
  ruijie-go info

Environment Variables:
  RUIJIE_USERNAME     Default username
  RUIJIE_PASSWORD     Default password
  RUIJIE_VERBOSE      Enable verbose output (1/true/yes)
  RUIJIE_SERVICE      Service name (default: 校园网)
  HTTP_PROXY          HTTP proxy URL
  HTTPS_PROXY         HTTPS proxy URL`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.ruijie-go.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	rootCmd.PersistentFlags().StringVar(&proxy, "proxy", "", "Proxy URL (e.g., socks5://127.0.0.1:1080)")

	// Bind flags to viper
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	viper.BindPFlag("proxy", rootCmd.PersistentFlags().Lookup("proxy"))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".ruijie-go" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".ruijie-go")
	}

	// Environment variables
	viper.SetEnvPrefix("RUIJIE")
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		if verbose {
			fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
		}
	}
}
