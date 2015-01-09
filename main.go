package main

import (
	log "github.com/Sirupsen/logrus"
)

func main() {
	log.SetLevel(log.DebugLevel)
	r := NewRegistry("http://registry.npmjs.org")

	log.Info(r.PackageVersions("express"))
}
