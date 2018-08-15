package main

import (
	"fmt"
	"log"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command {
	Use: "nudhcp",
	Short: "nudhcp is a simple and straight forward dhcp server",
	Long: `Who doesn't like YAML? nudhcp does! simple, straight forward, and yaml configurable!`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("welcome to nudhcp!")
	},
}


func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

