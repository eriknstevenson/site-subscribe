package main

import (
	"flag"
	"fmt"
	"os"

	ss "github.com/narrative/site-subscribe"
)

/*
usage:
site-subscribe sub <site> --url=<url> --selector=<selector> --user=<email address>
site-subscribe update --key=<sendgrid-key>
*/
func main() {

	subCommand := flag.NewFlagSet("sub", flag.ExitOnError)
	siteNameFlag := subCommand.String("name", "", "name of site")
	siteURLFlag := subCommand.String("url", "", "URL of site")
	selectorFlag := subCommand.String("selector", "body", "CSS selector to get content from site")
	userFlag := subCommand.String("user", "", "email address of user to notify of changes")

	newUserCommand := flag.NewFlagSet("new-user", flag.ExitOnError)
	userNameFlag := newUserCommand.String("name", "", "name of user")
	userEmailFlag := newUserCommand.String("email", "", "email address of user")

	updateCommand := flag.NewFlagSet("update", flag.ExitOnError)
	keyFlag := updateCommand.String("key", "", "SendGrid API key")

	displayUsage := func() {
		subCommand.Usage()
		newUserCommand.Usage()
		updateCommand.Usage()
	}

	if len(os.Args) == 1 {
		displayUsage()
		return
	}

	switch os.Args[1] {
	case "sub":
		subCommand.Parse(os.Args[2:])
		if !subCommand.Parsed() || *userFlag == "" || *siteNameFlag == "" || *siteURLFlag == "" {
			fmt.Println("Please provide required parameters.")
			displayUsage()
			return
		}
		ss.AddSite(*userFlag, ss.Site{Name: *siteNameFlag, URL: *siteURLFlag, ContentSelector: *selectorFlag})
	case "new-user":
		newUserCommand.Parse(os.Args[2:])
		if !newUserCommand.Parsed() || *userNameFlag == "" || *userEmailFlag == "" {
			fmt.Println("Please provide required parameters.")
			displayUsage()
			return
		}
		ss.AddUser(*userNameFlag, *userEmailFlag)
	case "update":
		updateCommand.Parse(os.Args[2:])
		if !updateCommand.Parsed() || *keyFlag == "" {
			fmt.Println("Please provide required parameters.")
			displayUsage()
			return
		}
		ss.Update(*keyFlag)

	default:
		displayUsage()
		return
	}
}
