/*
A cli application for using docker to quickly and easily manage infrastructure
and services through code.

Mach provides tooling around a simple docker and docker-machine based workflow for
managing infrastructure, services and docker images. This allows you to easily
leverage any git repository for the purposes of managing your Infrastructure as Code.
*/
package main

import (
	"github.com/spf13/viper"
	"github.com/superterran/mach/cmd"
)

func main() {

	viper.AddConfigPath(".")
	viper.SetEnvPrefix("mach")
	viper.AutomaticEnv()

	cmd.Execute()
}
