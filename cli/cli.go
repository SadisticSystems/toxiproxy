package main

import (
	"log"
	"sort"
	"strconv"
	"strings"

	"github.com/Shopify/toxiproxy/client"
	"github.com/codegangsta/cli"

	"fmt"
	"os"
)

const (
	redColor    = "\x1b[31m"
	greenColor  = "\x1b[32m"
	yellowColor = "\x1b[33m"
	blueColor   = "\x1b[34m"
	grayColor   = "\x1b[37m"
	noColor     = "\x1b[0m"
)

var toxicSymbols = map[string]string{
	"latency":    "L",
	"down":       "D",
	"bandwidth":  "B",
	"slow_close": "SC",
	"timeout":    "T",
}

func main() {
	defer fmt.Print(noColor) // make sure to clear unwanted colors
	toxiproxyClient := toxiproxy.NewClient("http://localhost:8474")
	fmt.Println("UNDER JAKE DEV")

	app := cli.NewApp()
	app.Name = "Toxiproxy"
	app.Usage = "Simulate network and system conditions"
	app.Commands = []cli.Command{
		{
			Name:    "list",
			Usage:   "list all proxies",
			Aliases: []string{"l", "li", "ls"},
			Action:  withToxi(list, toxiproxyClient),
		},
		{
			Name:    "inspect",
			Aliases: []string{"i", "ins"},
			Usage:   "inspect a single proxy",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "name, n",
					Usage: "name of the proxy",
				},
			},
			Action: withToxi(inspect, toxiproxyClient),
		},
		{
			Name:    "toggle",
			Usage:   "toggle enabled status on a proxy",
			Aliases: []string{"tog"},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "name, n",
					Usage: "name of the proxy",
				},
			},
			Action: withToxi(toggle, toxiproxyClient),
		},
		{
			Name:    "create",
			Usage:   "create a new proxy",
			Aliases: []string{"c", "new"},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "name, n",
					Usage: "name of the proxy",
				},
				cli.StringFlag{
					Name:  "listen, l",
					Usage: "proxy will listen on this address",
				},
				cli.StringFlag{
					Name:  "upstream, u",
					Usage: "proxy will forward to this address",
				},
			},
			Action: withToxi(create, toxiproxyClient),
		},
		{
			Name:    "delete",
			Usage:   "delete a proxy",
			Aliases: []string{"d"},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "name, n",
					Usage: "name of the proxy",
				},
			},
			Action: withToxi(delete, toxiproxyClient),
		},
		{
			Name:    "toxic",
			Aliases: []string{"t"},
			Usage:   "add, remove or update a toxic",
			Subcommands: []cli.Command{
				{
					Name:    "add",
					Aliases: []string{"a", "set", "s"},
					Usage:   "add a new toxic",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "name, n",
							Usage: "name of the proxy",
						},
						cli.StringFlag{
							Name:  "toxicName, tn",
							Usage: "name of the toxic",
						},
						cli.StringFlag{
							Name:  "type, t",
							Usage: "type of toxic",
						},
						cli.StringFlag{
							Name:  "fields, f",
							Usage: "comma seperated key=value toxic fields",
						},
						cli.BoolFlag{
							Name:  "upstream, u",
							Usage: "set toxic on upstream",
						},
						cli.BoolFlag{
							Name:  "downstream, d",
							Usage: "set toxic on downstream",
						},
					},
					Action: withToxi(addToxic, toxiproxyClient),
				},
				{
					Name:    "update",
					Aliases: []string{"u"},
					Usage:   "update an enabled toxic",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "name, n",
							Usage: "name of the proxy",
						},
						cli.StringFlag{
							Name:  "toxicName, tn",
							Usage: "name of the toxic",
						},
						cli.StringFlag{
							Name:  "fields, f",
							Usage: "comma seperated key=value toxic fields",
						},
					},
					Action: withToxi(updateToxic, toxiproxyClient),
				},
				{
					Name:    "remove",
					Aliases: []string{"r", "delete", "d"},
					Usage:   "remove an enabled toxic",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "name, n",
							Usage: "name of the proxy",
						},
						cli.StringFlag{
							Name:  "toxicName, tn",
							Usage: "name of the toxic",
						},
					},
					Action: withToxi(removeToxic, toxiproxyClient),
				},
			},
		},
	}

	app.Run(os.Args)
}

type toxiAction func(*cli.Context, *toxiproxy.Client)

func withToxi(f toxiAction, t *toxiproxy.Client) func(*cli.Context) {
	return func(c *cli.Context) {
		f(c, t)
	}
}

func list(c *cli.Context, t *toxiproxy.Client) {
	proxies, err := t.Proxies()
	if err != nil {
		log.Fatalln("Failed to retrieve proxies: ", err)
	}

	var proxyNames []string
	for proxyName := range proxies {
		proxyNames = append(proxyNames, proxyName)
	}
	sort.Strings(proxyNames)

	fmt.Fprintf(os.Stderr, "%sListen\t\t%sUpstream\t%sName%s\n", blueColor, yellowColor, greenColor, noColor)
	fmt.Fprintf(os.Stderr, "%s================================================================================%s\n", grayColor, noColor)

	for _, proxyName := range proxyNames {
		proxy := proxies[proxyName]
		fmt.Printf("%s%s\t%s%s\t%s%s%s\n", blueColor, proxy.Listen, yellowColor, proxy.Upstream, enabledColor(proxy.Enabled), proxy.Name, noColor)
	}
}
func inspect(c *cli.Context, t *toxiproxy.Client) {
	proxyName := getArgOrFail(c, "name")

	proxy, err := t.Proxy(proxyName)
	if err != nil {
		log.Fatalf("Failed to retrieve proxy %s: %s\n", proxyName, err.Error())
	}

	toxicPipe := formatToxicPipe(proxy.ActiveToxics)

	fmt.Printf("%s%s%s\n", enabledColor(proxy.Enabled), proxy.Name, noColor)
	fmt.Printf("%s%s %s--%s-> %s%s%s\n\n", blueColor, proxy.Listen, grayColor, toxicPipe, yellowColor, proxy.Upstream, noColor)

	listToxics(proxy.ActiveToxics, "upstream")
	fmt.Println()
	listToxics(proxy.ActiveToxics, "downstream")
}

func toggle(c *cli.Context, t *toxiproxy.Client) {
	proxyName := getArgOrFail(c, "name")

	proxy, err := t.Proxy(proxyName)
	if err != nil {
		log.Fatalf("Failed to retrieve proxy %s: %s\n", proxyName, err.Error())
	}

	proxy.Enabled = !proxy.Enabled

	err = proxy.Save()
	if err != nil {
		log.Fatalf("Failed to toggle proxy %s: %s\n", proxyName, err.Error())
	}

	fmt.Printf("Proxy %s%s%s is now %s%s%s\n", enabledColor(proxy.Enabled), proxyName, noColor, enabledColor(proxy.Enabled), enabledText(proxy.Enabled), noColor)
}

func create(c *cli.Context, t *toxiproxy.Client) {
	proxyName := getArgOrFail(c, "name")
	listen := getArgOrFail(c, "listen")
	upstream := getArgOrFail(c, "upstream")
	_, err := t.CreateProxy(proxyName, listen, upstream)
	if err != nil {
		log.Fatalf("Failed to create proxy: %s\n", err.Error())
	}
	fmt.Printf("Created new proxy %s\n", proxyName)
}

func delete(c *cli.Context, t *toxiproxy.Client) {
	proxyName := getArgOrFail(c, "name")
	p, err := t.Proxy(proxyName)
	if err != nil {
		log.Fatalf("Failed to retrieve proxy %s: %s\n", proxyName, err.Error())
	}

	err = p.Delete()
	if err != nil {
		log.Fatalf("Failed to delete proxy: %s\n", err.Error())
	}
	fmt.Printf("Deleted proxy %s\n", proxyName)
}

func addToxic(c *cli.Context, t *toxiproxy.Client) {
	proxyName := getArgOrFail(c, "name")
	toxicName := c.String("toxicName")
	toxicType := getArgOrFail(c, "type")
	toxicFields := getArgOrFail(c, "fields")

	upstream := c.Bool("upstream")
	downstream := c.Bool("downstream")

	fields := parseFields(toxicFields)

	p, err := t.Proxy(proxyName)
	if err != nil {
		log.Fatalf("Failed to retrieve proxy %s: %s\n", proxyName, err.Error())
	}

	addToxic := func(stream string) {
		_, err := p.AddToxic(toxicName, toxicType, stream, fields)
		if err != nil {
			log.Fatalf("Failed to set toxic: %s\n", err.Error())
		}
		fmt.Printf("Set %s %s toxic '%s' on proxy '%s'\n", stream, toxicType, toxicName, proxyName)
	}

	if upstream {
		addToxic("upstream")
	}
	// Default to downstream.
	if downstream || (!downstream && !upstream) {
		addToxic("downstream")
	}
}

func updateToxic(c *cli.Context, t *toxiproxy.Client) {
	proxyName := getArgOrFail(c, "name")
	toxicName := getArgOrFail(c, "toxicName")
	toxicFields := getArgOrFail(c, "fields")

	fields := parseFields(toxicFields)

	p, err := t.Proxy(proxyName)
	if err != nil {
		log.Fatalf("Failed to retrieve proxy %s: %s\n", proxyName, err.Error())
	}

	_, err = p.UpdateToxic(toxicName, fields)
	if err != nil {
		log.Fatal("Failed to update toxic: %s\n", err.Error())
	}

	fmt.Printf("Updated toxic '%s' on proxy '%s'\n", toxicName, proxyName)
}

func removeToxic(c *cli.Context, t *toxiproxy.Client) {
	proxyName := getArgOrFail(c, "name")
	toxicName := getArgOrFail(c, "toxicName")

	p, err := t.Proxy(proxyName)
	if err != nil {
		log.Fatalf("Failed to retrieve proxy %s: %s\n", proxyName, err.Error())
	}

	err = p.RemoveToxic(toxicName)
	if err != nil {
		log.Fatal("Failed to remove toxic: %s\n", err.Error())
	}

	fmt.Printf("Removed toxic '%s' on proxy '%s'\n", toxicName, proxyName)
}

func parseFields(raw string) toxiproxy.Toxic {
	parsed := map[string]interface{}{}
	keyValues := strings.Split(raw, ",")
	for _, keyValue := range keyValues {
		kv := strings.SplitN(keyValue, "=", 2)
		if len(kv) != 2 {
			log.Fatal("Fields should be in format key1=value1,key2=value2\n")
		}
		value, err := strconv.Atoi(kv[1])
		if err != nil {
			log.Fatal("Toxic field was expected to be an integer.\n")
		}
		parsed[kv[0]] = value
	}
	return parsed
}

func enabledColor(enabled bool) string {
	if enabled {
		return greenColor
	}

	return redColor
}

func enabledText(enabled bool) string {
	if enabled {
		return "enabled"
	}

	return "disabled"
}

func listToxics(toxics toxiproxy.Toxics, stream string) {
	fmt.Printf("%s%s\n", noColor, stream)
	for name, toxic := range toxics {
		if toxic["stream"].(string) != stream {
			continue
		}
		fmt.Printf("%s%s stream=%s", greenColor, name, stream)
		for property, value := range toxic {
			fmt.Printf(" %s=", property)
			fmt.Print(value)
		}
		fmt.Println()
	}
}

func getArgOrFail(c *cli.Context, name string) string {
	arg := c.String(name)
	if arg == "" {
		log.Fatalf("Required argument '%s' was empty.\n", name)
	}
	return arg
}

func formatToxicPipe(toxics toxiproxy.Toxics) string {
	var pipe string
	for _, t := range toxics {
		kind := t["type"].(string)
		s, ok := toxicSymbols[kind]
		if !ok {
			log.Printf("Cannot find symbol for '%s' toxic\n", kind)
		}
		pipe += s + "-"
	}
	return pipe
}
