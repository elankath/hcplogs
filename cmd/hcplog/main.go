package main

import (
	"flag"
	"fmt"
	"github.com/elankath/hcptool/pkg/hcplog"
	"log"
	"os"
	"os/user"
	"strings"
	"syscall"

	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh/terminal"
)

// type Person struct {
// 	Name string
// 	Age int
// }
//
// func (p *Person) String() string {
// 	return fmt.Sprintf("(%s,%d)", p.Name, p.Age)
// }
//
// func main() {
// 	people := []Person{ Person{"Madhav", 39 }, Person{"Tarun", 38}}
// 	fmt.Printf("people with percent-v = %v\n", people)
// 	fmt.Printf("people with percent-s= %s\n", people)
// }

// Params represents command line parameters

var commands = []string{"list", "grab"}
func main() {
	log.SetFlags(0)
	log.SetPrefix("hcplog: ")
	var config hcplog.Config
	u, err := user.Current()
	if err != nil {
		log.Fatal(errors.New("Can't determine current u"))
	}
	coord := flag.String("c", "", "HCP App Coordinate in the form account:application")
	flag.StringVar(&config.User, "u", u.Username, "HCP User")
	flag.StringVar(&config.Password, "p", "", "HCP Password")

	flag.StringVar(&config.LandscapeHost, "l", "", "HCP Landscape Host")

	// listCommand := flag.NewFlagSet("list", flag.ExitOnError)
	// grabCommand := flag.NewFlagSet("grab", flag.ExitOnError)

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
		println("No subcommand specified!")
		printUsage()
		os.Exit(2)
	} else {
		command  = args[0]
	}
	// switch os.Args[1] {
	// case "list":
	// 	askCommand.Parse(os.Args[2:])
	// case "send":
	// 	sendCommand.Parse(os.Args[2:])
	// default:
	// 	fmt.Printf("%q is not valid command.\n", os.Args[1])
	// 	printUsage()
	// 	os.Exit(2)
	// }
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
	if strings.EqualFold(command, "list") {
		if err := client.PrintFiles(application, os.Stdout); err != nil {
			log.Fatal(err)
		}
	} else if strings.EqualFold(command, "grab") {
		if len(args) < 2 {
			log.Fatal(errors.New("grab requires at-least one log file name or glob pattern as argument"))
		}
		names := args[1:]
		if err := client.GrabFilesAndPrint(application, names, os.Stdout); err != nil {
			log.Fatal(err)
		}
	} else {
		fmt.Printf("%q is not valid command.\n", os.Args[1])
		printUsage()
		os.Exit(2)
	}
}

func printUsage() {
	println()
	println("USAGE:")
	println("hcplog -c <account:application> -u <user> -p <password> -l <landscapeHost> list")
	println("hcplog -c <account:application> -u <user> -p <password> -l <landscapeHost> grab <fileGlobPattern>")
	println()
	println("NOTES")
	println("* password need not be passed as option and can be entered in standard input")
	println("* TOOD: Introduce parallel download with -c (concurrent) option ")
	println("* TOOD: Specify output dir for downlaod with -o option ")
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
