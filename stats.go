package main

import (
	"encoding/csv"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
)

func stats() error {
	pkgs, err := packageNames()
	if err != nil {
		return err
	}

	valid := 0
	for _, p := range pkgs {
		filename := filepath.Join("meta", p+".json")
		if _, err := os.Stat(filename); err == nil {
			valid += 1
		}
	}

	var py3 []string
	blob, err := ioutil.ReadFile("python3-packages.json")
	if err != nil {
		return err
	}
	if err := json.Unmarshal(blob, &py3); err != nil {
		return err
	}

	var results map[string]result
	var broken int
	var errors int
	blob, err = ioutil.ReadFile("results.json")
	if err != nil {
		return err
	}
	if err := json.Unmarshal(blob, &results); err != nil {
		return err
	}

	output := []result{}
	dead := map[string]int{}

	for name, r := range results {
		if len(r.Errors) > 0 {
			errors += 1
		}
		if len(r.Imports) == 0 {
			continue
		}

		broken += 1

		r.Package = name
		output = append(output, r)

		for battery, _ := range r.Imports {
			dead[battery] += 1
		}
	}

	log.Printf("  total packages: %d", len(pkgs))
	log.Printf("  valid packages: %d", valid)
	log.Printf("python3 packages: %d", len(py3))
	log.Printf("scanned packages: %d", len(results))
	log.Printf(" broken packages: %d", broken)
	log.Printf("errored packages: %d", errors)

	// Write out package information
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
