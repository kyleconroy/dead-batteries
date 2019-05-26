package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"sort"
	"strings"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

type Info struct {
	Package string   `json:"project"`
	Version string   `json:"version"`
	URL     string   `json:"url"`
	Imports []string `json:"imports"`
}

func run() error {
	var results map[string]map[string]int
	blob, err := ioutil.ReadFile("results-may-25.json")
	if err != nil {
		return err
	}
	if err := json.Unmarshal(blob, &results); err != nil {
		return err
	}

	pkgs := map[string]Info{}

	for name, imports := range results {
		oparts := strings.SplitN(name, "/", 2)
		pkgver := oparts[0]

		switch {
		case strings.HasSuffix(pkgver, ".whl"):
			a := strings.Split(pkgver, "-")
			a = a[:len(a)-1] // Pop off any.whl
			a = a[:len(a)-1] // Pop off none
			a = a[:len(a)-1] // Pop off py*
			pkgver = strings.Join(a, "-")
		case strings.HasSuffix(pkgver, ".tar.gz"):
			pkgver = strings.Replace(pkgver, ".tar.gz", "", 1)
		case strings.HasSuffix(pkgver, ".egg"):
			a := strings.Split(pkgver, "-")
			a = a[:len(a)-1] // Pop off py*.egg
			pkgver = strings.Join(a, "-")
		case strings.HasSuffix(pkgver, ".zip"):
			pkgver = strings.Replace(pkgver, ".zip", "", 1)
		default:
			panic("unknown format: " + pkgver)
		}

		a := strings.Split(pkgver, "-")
		var pkg, version string

		if len(a) == 1 {
			pkg = pkgver
		} else {
			version, a = a[len(a)-1], a[:len(a)-1]
			pkg = strings.Join(a, "-")
		}

		// Underscores and dashes are treated as the same in package names
		pkg = strings.Replace(pkg, "_", "-", -1)

		if _, ok := pkgs[pkg]; ok {
			continue
		}
		if len(imports) == 0 {
			continue
		}

		found := map[string]string{}
		for battery, _ := range imports {
			found[strings.TrimSpace(battery)] = ""
		}

		keys := make([]string, 0, len(found))
		for k := range found {
			keys = append(keys, k)
		}

		pkgs[pkg] = Info{
			Package: pkg,
			Version: version,
			URL:     fmt.Sprintf("https://pypi.org/project/%s/%s/", pkg, version),
			Imports: keys,
		}
	}

	output := []Info{}
	for _, i := range pkgs {
		output = append(output, i)
	}
	sort.Slice(output, func(i, j int) bool { return output[i].Package < output[j].Package })

	blob, err = json.MarshalIndent(&output, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile("packages.json", blob, 0644)
}
