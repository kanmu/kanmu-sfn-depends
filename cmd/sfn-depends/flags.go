package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/k1LoW/duration"
)

var (
	version string
)

type flags struct {
	stateMachineArns []string
	period           time.Duration
}

func init() {
	cmdLine := flag.NewFlagSet(filepath.Base(os.Args[0]), flag.ExitOnError)

	cmdLine.Usage = func() {
		fmt.Fprintf(cmdLine.Output(), "Usage: %s [OPTION] STATE_MACHINE_ARN...\n", cmdLine.Name())
		cmdLine.PrintDefaults()
	}

	flag.CommandLine = cmdLine
}

func parseFlags() (*flags, error) {
	var periodStr string

	flag.StringVar(&periodStr, "period", "", "period")
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Parse()

	if *showVersion {
		printVersionAndExit()
	}

	args := flag.Args()

	if len(args) < 1 {
		printUsageAndExit()
	}

	if periodStr == "" {
		log.Fatal("'-period' is required")
	}

	period, err := duration.Parse(periodStr)

	if err != nil {
		log.Fatal(err)
	}

	f := &flags{
		stateMachineArns: args,
		period:           period,
	}

	return f, nil
}

func printVersionAndExit() {
	v := version

	if v == "" {
		v = "<nil>"
	}

	fmt.Fprintln(flag.CommandLine.Output(), v)
	os.Exit(0)
}

func printUsageAndExit() {
	flag.CommandLine.Usage()
	os.Exit(0)
}
