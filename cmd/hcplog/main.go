package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/user"
	"strings"
	"syscall"

	"github.com/elankath/hcptool/pkg/log"
	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh/terminal"
)

var commands = map[string]bool{"list": true, "grab": true, "set": true}

// Params represents command line parameters
func main() {
	log.SetFlags(0)
	log.SetPrefix("hcplog: ")
	var config hcplog.Config
	user, err := user.Current()
	if err != nil {
		log.Fatal(errors.New("Can't determine current user"))
	}
	coord := flag.String("c", "", "HCP App Coordinate in the form account:application")
	flag.StringVar(&config.User, "u", user.Username, "HCP User")
	flag.StringVar(&config.Password, "p", "", "HCP Password")

	flag.StringVar(&config.LandscapeHost, "l", "", "HCP Landscape Host")
	flag.Parse()

	if strings.TrimSpace(*coord) == "" {
		log.Fatal(errors.New("HCP app coordinate must be specified via -c in the form account:application"))
	}

	splits := strings.Split(*coord, ":")
	if len(splits) != 2 {
		log.Fatal(errors.Errorf("Invalid coordinate: '%s'. Coordinate must be in form account:application", *coord))
	}
	config.Account = splits[0]
	application := splits[1]

	if config.Account == "" || application == "" {
		log.Fatal(errors.Errorf("Cannot parse account or application from coordinate: '%s'. Account and Application must be present and must be part of coordinate in the form account:application", *coord))
	}

	if config.LandscapeHost == "" {
		log.Fatal(errors.New("HCP landscape host must be specified via -l"))
	}

	var command string
	args := flag.Args()
	if len(args) == 0 {
		command = "list"
	} else if !commands[args[0]] {
		log.Fatal(errors.Errorf("Unknown command: %s", args[0]))
	} else {
		command = args[0]
	}

	if config.Password == "" {
		config.Password, err = readPassword()
		if err != nil {
			log.Fatal(err)
		}
	}

	client, err := hcplog.NewClient(config)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Executing'%s' command using config: %s\n", command, &config)
	if err := client.PrintFiles(application, os.Stdout); err != nil {
		log.Fatal(err)
	}
}

func readPassword() (string, error) {
	fmt.Print("Enter Password: ")
	bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", errors.Wrap(err, "Cannot read password from standard input")
	}
	fmt.Println()
	return string(bytePassword), nil
}
