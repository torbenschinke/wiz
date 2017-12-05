package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
)

const VERSION = "version 3.0.0 alpha (Mark III)"

type StringValue struct {
	Value string
}

func (s *StringValue) Set(value string) error {
	s.Value = value
	return nil
}

// Implement the flag.Value interface
func (s *StringValue) String() string {
	return fmt.Sprintf("%v", *s)
}

var Usage = func() {
	fmt.Fprintf(os.Stderr, "Usage of wiz:\n")
	flag.PrintDefaults()
}

func main() {
	parse()

	flag.Usage = Usage
	cmdInit := flag.Bool("init", false, "Create an empty Wiz repository in the current working directory")
	cmdVersion := flag.Bool("version", false, "Prints current Wiz cmdVersion and exists")
	flag.Parse()

	if *cmdVersion {
		printVersion()
		os.Exit(0)
	}

	if *cmdInit {
		createEmptyRepo()
		os.Exit(0)
	}

	Usage()
}

func printVersion() {
	fmt.Fprintf(os.Stdout, "wiz %s\n", VERSION)
	fmt.Fprintf(os.Stdout, "%s on %s/%s (%d cores)\n", runtime.Version(), runtime.GOOS, runtime.GOARCH, runtime.NumCPU())
}

func createEmptyRepo() {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	dir = filepath.Join(dir, ".wiz")
	_, err = os.Stat(dir)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Fatal(err)
		}
	} else {
		log.Fatal("Repository already exist, cannot init non empty directory in", dir)
	}
	fmt.Fprintf(os.Stderr, "Initializing empty Wiz repository in %s\n", dir)

}
