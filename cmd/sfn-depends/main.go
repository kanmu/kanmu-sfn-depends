package main

import (
	"log"

	sfndepents "github.com/kanmu/sfn-depends"
)

func init() {
	log.SetFlags(0)
}

func main() {
	flags, err := parseFlags()

	if err != nil {
		log.Fatal(err)
	}

	client, err := sfndepents.NewClient()

	if err != nil {
		log.Fatal(err)
	}

	err = client.Validate(flags.stateMachineArns, flags.period)

	if err != nil {
		log.Fatal(err)
	}
}
