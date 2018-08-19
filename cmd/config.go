package main

import (
	"log"
	"github.com/jtaylorcpp/nudhcp"
	"github.com/spf13/cobra"
)

var configFile string

func init() {
	runCmd.Flags().StringVarP(&configFile, "config-file","f","","Config file to run from")

	rootCmd.AddCommand(runCmd)
}

var runCmd = &cobra.Command {
	Use: "run",
	Short: "run a nudhcp server",
	Run: func(cmd *cobra.Command, args []string) {
		log.Println("starting dhcp servers from file: ",configFile)
		dhcpManager := nudhcp.LoadFromFile(configFile)
		log.Println("dhcp servers: ",dhcpManager)
		dhcpManager.StartDHCPServers()
	},
}
