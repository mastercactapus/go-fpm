package main

import (
	log "github.com/Sirupsen/logrus"
	"os"
)

func main() {
	log.SetLevel(log.DebugLevel)
	r := NewRegistry("http://registry.npmjs.org")
	m := make(DependencyMap, 2)
	var err error
	req, err := NewSemverRequirements("^4")
	if err != nil {
		log.Fatalln(err)
	}
	m["express"] = req
	req, err = NewSemverRequirements("")
	if err != nil {
		log.Fatalln(err)
	}
	m["grunt"] = req

	tree, err := CalculateTree(r, m)
	if err != nil {
		log.Fatalln(err)
	}
	tree.Print(os.Stdout)
}
