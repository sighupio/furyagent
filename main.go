package main

import (
	"fmt"
	"log"

	"github.com/spf13/viper"
)

func main() {
	viper.SetConfigFile(configFile)
	viper.SetConfigType("yml")
	viper.AddConfigPath(".")
	configuration := new(Furyconf)
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}
	err := viper.Unmarshal(configuration)
	if err != nil {
		log.Fatalf("unable to decode into struct, %v", err)
	}
	fmt.Println(configuration)
	configuration.Validate()
	fmt.Println("starting download")
	log.Println(configuration.Download())
}
