package main

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/chzyer/readline"
	"github.com/vishvananda/netlink"
)

type CommandNode struct {
	Children map[string]*CommandNode
	Dynamic  func() []string
}

var rootCommand = &CommandNode{
	Children: map[string]*CommandNode{
		"show": {
			Children: map[string]*CommandNode{
				"link":    {},
				"address": {},
			},
		},
		"set": {
			Children: map[string]*CommandNode{
				"interface": {
					Dynamic: getInterfaces,
				},
			},
		},
		"exit": {},
		"quit": {},
	},
}

func getInterfaces() []string {
	links, _ := netlink.LinkList()
	var names []string
	for _, link := range links {
		names = append(names, link.Attrs().Name)
	}
	return names
}

type completer struct{}

func (c *completer) Do(line []rune, pos int) ([][]rune, int) {
	input := string(line[:pos])
	fields := strings.Fields(input)
	prefix := ""
	if len(fields) > 0 && !strings.HasSuffix(input, " ") {
		prefix = fields[len(fields)-1]
		fields = fields[:len(fields)-1]
	}

	suggestions := getSuggestions(fields, rootCommand)
	var out [][]rune
	for _, s := range suggestions {
		if strings.HasPrefix(s, prefix) {
			out = append(out, []rune(s))
		}
	}
	return out, len(prefix)
}

func getSuggestions(path []string, node *CommandNode) []string {
	if node == nil {
		return nil
	}
	if len(path) == 0 {
		var out []string
		for k := range node.Children {
			out = append(out, k)
		}
		if node.Dynamic != nil {
			out = append(out, node.Dynamic()...)
		}
		return out
	}

	next := path[0]

	// Bei Dynamic prüfen wir, ob z. B. eth0 existiert
	if node.Dynamic != nil {
		for _, opt := range node.Dynamic() {
			if opt == next {
				return getSuggestions(path[1:], &CommandNode{})
			}
		}
	}

	child, ok := node.Children[next]
	if !ok {
		return nil
	}
	return getSuggestions(path[1:], child)
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

	rl, err := readline.NewEx(&readline.Config{
		Prompt:          "netshell> ",
		AutoComplete:    &completer{},
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
		HistoryFile:     "/tmp/netshell.history",
	})
	if err != nil {
		panic(err)
	}
	defer rl.Close()

	for {
		line, err := rl.Readline()
		if err != nil {
			break
		}
		cmd := strings.TrimSpace(line)

		switch cmd {
			case "exit", "quit":
				fmt.Println("Bye!")
				return
			case "show link":
				showLinks()
			case "show address":
				showAddresses()
			default:
				if cmd != "" {
					fmt.Println("Unknown command:", cmd)
				}
		}
	}
}
