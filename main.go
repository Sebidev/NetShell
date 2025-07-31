package main

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/c-bata/go-prompt"
	"github.com/vishvananda/netlink"
)

func executor(in string) {
	in = strings.TrimSpace(in)
	switch in {
		case "exit", "quit":
			fmt.Println("Bye!")
			os.Exit(0)

		case "show link":
			showLinks()

		case "show address":
			showAddresses()

		default:
			fmt.Println("Unknown command:", in)
	}
}

func completer(d prompt.Document) []prompt.Suggest {
	return []prompt.Suggest{
		{Text: "show link", Description: "Show Interfaces"},
		{Text: "show address", Description: "Show IP-Adressen"},
		{Text: "exit", Description: "Leave Shell"},
		{Text: "quit", Description: "Leave Shell"},
	}
}

func showLinks() {
	links, _ := netlink.LinkList()
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	fmt.Fprintln(w, "Name\tType\tState")

	for _, link := range links {
		attrs := link.Attrs()
		state := "DOWN"
		if attrs.OperState.String() == "up" {
			state = "UP"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\n", attrs.Name, link.Type(), state)
	}

	w.Flush()
}

func showAddresses() {
	links, _ := netlink.LinkList()
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	fmt.Fprintln(w, "Interface\tAddress")

	for _, link := range links {
		addrs, _ := netlink.AddrList(link, netlink.FAMILY_ALL)
		for _, addr := range addrs {
			fmt.Fprintf(w, "%s\t%s\n", link.Attrs().Name, addr.IPNet.String())
		}
	}

	w.Flush()
}


func main() {
	fmt.Println("Welcome to NetShell!")
	p := prompt.New(
		executor,
		 completer,
		 prompt.OptionPrefix("netshell> "),
			prompt.OptionTitle("NetShell CLI"),
	)
	p.Run()
}
