package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type Info struct {
	Package string   `json:"project"`
	Version string   `json:"version"`
	URL     string   `json:"url"`
	Imports []string `json:"imports"`
}

func trimFormat(pkgver string) string {
	switch {
	case strings.HasSuffix(pkgver, ".whl"):
		a := strings.Split(pkgver, "-")
		a = a[:len(a)-1] // Pop off any.whl
		a = a[:len(a)-1] // Pop off none
		a = a[:len(a)-1] // Pop off py*
		return strings.Join(a, "-")
	case strings.HasSuffix(pkgver, ".tar.gz"):
		return strings.Replace(pkgver, ".tar.gz", "", 1)
	case strings.HasSuffix(pkgver, ".egg"):
		a := strings.Split(pkgver, "-")
		a = a[:len(a)-1] // Pop off py*.egg
		return strings.Join(a, "-")
	case strings.HasSuffix(pkgver, ".zip"):
		return strings.Replace(pkgver, ".zip", "", 1)
	default:
		panic("unknown format: " + pkgver)
	}
}

func split(input string) (string, string) {
	input = strings.Replace(input, ".linux-x86_64", "", -1)

	a := strings.Split(input, "-")
	var pkg, version string

	if len(a) == 1 {
		pkg = input
	} else if len(a) == 2 {
		pkg = a[0]
		version = a[1]
	} else {
		for i, part := range a {
			// The version can never be the first part
			if i == 0 {
				continue
			}
			containsNumber, _ := regexp.MatchString(`[0-9]+`, part)
			// containsNumberDot := containsNumber && strings.Contains(part, ".")

			_, err := strconv.Atoi(part)
			containsAllNumbers := err == nil

			if containsNumber || containsAllNumbers {
				pkg = strings.Join(a[:i], "-")
				version = strings.Join(a[i:], "-")
				break
			}
		}

		if pkg == "" && a[len(a)-1] == "dev" {
			pkg = strings.Join(a[:len(a)-1], "-")
			version = "dev"
		}

		if pkg == "" {
			pkg = strings.Join(a, "-")
		}
	}

	// Underscores and dashes are treated as the same in package names
	pkg = strings.Replace(pkg, "_", "-", -1)
	pkg = strings.Replace(pkg, " ", "-", -1)
	return pkg, version
}

func genpkg() error {
	var results map[string]map[string]int
	blob, err := ioutil.ReadFile("results.json")
	if err != nil {
		return err
	}
	if err := json.Unmarshal(blob, &results); err != nil {
		return err
	}

	pkgs := map[string]Info{}
	dead := map[string]int{}

	for name, imports := range results {
		oparts := strings.SplitN(name, "/", 2)
		pkgver := trimFormat(oparts[0])
		pkg, version := split(pkgver)

		if pkg == "" {
			log.Printf("parser err: %s", oparts[0])
			continue
		}
		if _, ok := pkgs[pkg]; ok {
			continue
		}
		if len(imports) == 0 {
			continue
		}

		found := map[string]string{}
		for battery, _ := range imports {
			found[strings.TrimSpace(battery)] = ""
			dead[strings.TrimSpace(battery)] += 1
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

	// Write out package information
	output := []Info{}
	for _, i := range pkgs {
		output = append(output, i)
	}
	sort.Slice(output, func(i, j int) bool { return output[i].Package < output[j].Package })

	blob, err = json.MarshalIndent(&output, "", "  ")
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile("packages.json", blob, 0644); err != nil {
		return err
	}

	// Write out imports information
	f, err := os.Create("imports.csv")
	if err != nil {
		return err
	}
	w := csv.NewWriter(f)
	if err := w.Write([]string{"package", "imports"}); err != nil {
		return err
	}

	for pkg, count := range dead {
		if err := w.Write([]string{pkg, strconv.Itoa(count)}); err != nil {
			return err
		}
	}

	w.Flush()
	return w.Error()
}
