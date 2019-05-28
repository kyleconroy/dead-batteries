package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
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
	for _, r := range results {
		if len(r.Errors) > 0 {
			errors += 1
		}
		if len(r.Imports) > 0 {
			broken += 1
		}
	}

	log.Printf("  total packages: %d", len(pkgs))
	log.Printf("  valid packages: %d", valid)
	log.Printf("python3 packages: %d", len(py3))
	log.Printf("scanned packages: %d", len(results))
	log.Printf(" broken packages: %d", broken)
	log.Printf("errored packages: %d", errors)
	return nil
}
