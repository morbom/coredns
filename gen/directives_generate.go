package main

import (
	"bufio"
	"io/ioutil"
	"os"
	"sort"
	"strconv"
	"strings"
)

var middlewarePath = "github.com/miekg/coredns/middleware/"
var header = "// generated by directives_generate.go; DO NOT EDIT\n"

func main() {
	mwFile := os.Args[1]

	mi := make(map[string]string, 0)
	md := make(map[int]string, 0)

	if file, err := os.Open(mwFile); err == nil {
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			if !strings.HasPrefix(line, `//`) && !strings.HasPrefix(line, "#") {
				items := strings.Split(line, ":")
				if len(items) == 3 {
					if priority, err := strconv.Atoi(items[0]); err == nil {
						md[priority] = items[1]
					}

					if items[2] != "" {
						if strings.Contains(items[2], "/") {
							mi[items[1]] = items[2]
						} else {
							mi[items[1]] = middlewarePath + items[2]
						}
					}

				}
			}
		}

		var orders []int
		for k := range md {
			orders = append(orders, k)
		}
		sort.Ints(orders)

		if os.Getenv("GOPACKAGE") == "core" {
			genImports("zmiddleware.go", mi)
		}
		if os.Getenv("GOPACKAGE") == "dnsserver" {
			genDirectives("zdirectives.go", md)
		}

	} else {
		os.Exit(1)
	}
	os.Exit(0)

}

func genImports(file string, mi map[string]string) {
	outs := header + "package " + os.Getenv("GOPACKAGE") + "\n\n" + "import ("

	if len(mi) > 0 {
		outs += "\n"
	}

	for _, v := range mi {
		outs += "		_ \"" + v + "\"\n"
	}
	outs += ")\n"

	err := ioutil.WriteFile(file, []byte(outs), 0644)
	if err != nil {
		os.Exit(1)
	}

}

func genDirectives(file string, md map[int]string) {

	outs := header + "package " + os.Getenv("GOPACKAGE") + "\n\n"
	outs += `
// Directives are registered in the order they should be
// executed.
//
// Ordering is VERY important. Every middleware will
// feel the effects of all other middleware below
// (after) them during a request, but they must not
// care what middleware above them are doing.

var directives = []string{
`

	var orders []int
	for k := range md {
		orders = append(orders, k)
	}
	sort.Ints(orders)

	for _, k := range orders {
		outs += "		\"" + md[k] + "\",\n"
	}

	outs += "}\n"

	err := ioutil.WriteFile(file, []byte(outs), 0644)
	if err != nil {
		os.Exit(1)
	}
}
