package main

import (
	"oci-collector/cmd"
	"oci-collector/util"
)

func main() {
	util.PrintBanner()
	cmd.Execute()
}