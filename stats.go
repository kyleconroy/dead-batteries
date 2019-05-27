package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

func stats() error {
	var packages []Info
	blob, err := ioutil.ReadFile("packages.json")
	if err != nil {
		return err
	}
	if err := json.Unmarshal(blob, &packages); err != nil {
		return err
	}
	var py3 int
	for _, p := range packages {
		url := fmt.Sprintf("https://pypi.python.org/pypi/%s/%s/json", p.Package, p.Version)
		resp, err := http.Get(url)
		if err != nil {
			return fmt.Errorf("%s %s", url, err)
		}
		if resp.StatusCode == http.StatusNotFound {
			log.Printf("package=%s version=%s err=not-found", p.Package, p.Version)
			continue
		}
		defer resp.Body.Close()
		var pypi PyPI
		dec := json.NewDecoder(resp.Body)
		if err := dec.Decode(&pypi); err != nil {
			log.Printf("package=%s version=%s err=%s", p.Package, p.Version, err)
			continue
			// return fmt.Errorf("decode error %s: %s", p.Package, err)
		}
		if pypi.SupportsPython3() {
			py3 += 1
		}
	}
	log.Printf("Total packages: %d", len(packages))
	log.Printf("Python 3 packages: %d", py3)
	return nil
}
