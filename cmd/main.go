package main

import (
	"fmt"
	"github.com/M-Ro/aurora-ai/cmd/instance"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
)

var rootCmd = &cobra.Command{}

func init() {
	rootCmd.AddCommand(instance.NewCmd())
}

// initialises viper config library.
func initConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	err := viper.ReadInConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Println("Loaded conf:", viper.ConfigFileUsed())
}

func main() {
	initConfig()

	err := rootCmd.Execute()
	if err != nil {
		fmt.Println(os.Stderr, err)
		os.Exit(1)
	}
}
