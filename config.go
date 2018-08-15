package nudhcp

import (
	"io/ioutil"
	"log"
	"time"
	"gopkg.in/yaml.v2"
)

// ConfigFile is the config for a set of dhcp servers, exporters, and API's
type ConfigFile struct {
	Servers []ServerConfig `yaml:"servers"`
}

// ServerConfig is the config for an individual dhcp server
type ServerConfig struct {
	Interface string `yaml:"interface"`
	ServerAddress string `yaml:"serverAddress"`
	Subnet string `yaml:"subnet"`
	Gateway string `yaml:"gateway"`
	DNS string `yaml:"dnsServers"`
	Duration string `yaml:"leaseDuration"`
	IPReservations []IPReservation `yaml:"ipReservations"`
}

// IPReservation which can be used to set "dynamic" static ip addresses
type IPReservation struct {
	Mac string `yaml:"mac"`
	IP string `yaml:"ip"`
}


// Parse a already read yaml file into the ConfigFile
func parseConfigFile(file string) *ConfigFile {
	newConfig := &ConfigFile{}
	err := yaml.Unmarshal([]byte(file), newConfig)
	if err != nil {
		log.Fatal("Unable to parse provided config file")
	}
	return newConfig
}

// Parse a ConfigFile Server
func serverFromConfig(sc ServerConfig) *DHCPServer {
	timeDuration,err := time.ParseDuration(sc.Duration)
	if err != nil {
		log.Fatal("Unable to parse time duration, please look at the godoc for time.ParseDuration: ",sc.Duration)
	}
	return NewDHCPServer(sc.Interface, sc.ServerAddress,
		sc.Subnet, sc.Gateway, sc.DNS,
		timeDuration)

}

// Load a yaml file and parse out all DHCP servers
func LoadFromFile(filePath string) *DHCPManager {
	yamlFile, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Fatal("Unable to parse yaml file")
	}

	newConfig := &ConfigFile{}
	err = yaml.Unmarshal(yamlFile, newConfig)

	newManager := &DHCPManager{servers: make(map[string]*DHCPServer)}
	for _,server := range newConfig.Servers {
		newManager.servers[server.Interface] = serverFromConfig(server)
	}

	return newManager
}

