package main

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

type Parser struct {
	lock    sync.RWMutex
	wg      sync.WaitGroup
	in      chan string
	results map[string]map[string]int
}

func (p *Parser) Work() {
	p.wg.Add(1)
	defer p.wg.Done()
	for {
		select {
		case path, ok := <-p.in:
			if !ok {
				return
			}
			if err := p.Unpack(path); err != nil {
				log.Printf("result=failure package=%s err=%s\n", filepath.Base(path), err)
			} else {
				log.Printf("result=success package=%s\n", filepath.Base(path))
			}
		}
	}
}

func (p *Parser) Unpack(pkg string) error {
	d, name := filepath.Split(pkg)
	_, hash := filepath.Split(filepath.Dir(d))
	key := name + "/" + hash

	p.lock.RLock()
	_, ok := p.results[key]
	p.lock.RUnlock()
	if ok {
		return nil
	}

	results := map[string]int{}

	if strings.HasSuffix(pkg, ".tar.gz") {
		r, err := os.Open(pkg)
		if err != nil {
			return err
		}
		defer r.Close()

		gzf, err := gzip.NewReader(r)
		if err != nil {
			return err
		}

		tr := tar.NewReader(gzf)
		for {
			hdr, err := tr.Next()
			if err == io.EOF {
				break // End of archive
			}
			if err != nil {
				return err
			}
			if !strings.HasSuffix(hdr.Name, ".py") {
				continue
			}
			tmpfile, err := ioutil.TempFile("", "pep594")
			if err != nil {
				return err
			}
			_, err = io.Copy(tmpfile, tr)
			if err != nil {
				return err
			}
			if err := tmpfile.Close(); err != nil {
				return err
			}
			p.Parse(results, tmpfile.Name())
		}
	} else if strings.HasSuffix(pkg, ".whl") || strings.HasSuffix(pkg, ".zip") {
		r, err := zip.OpenReader(pkg)
		if err != nil {
			return err
		}
		defer r.Close()
		for _, f := range r.File {
			if !strings.HasSuffix(f.Name, ".py") {
				continue
			}
			rc, err := f.Open()
			if err != nil {
				return err
			}
			defer rc.Close()
			tmpfile, err := ioutil.TempFile("", "pep594")
			if err != nil {
				return err
			}
			_, err = io.Copy(tmpfile, rc)
			if err != nil {
				return err
			}
			if err := tmpfile.Close(); err != nil {
				return err
			}
			p.Parse(results, tmpfile.Name())
		}
	} else {
		fmt.Println("unknown", pkg)
	}

	p.lock.Lock()
	p.results[key] = results
	p.lock.Unlock()

	return nil
}

func (p *Parser) Save() error {
	p.lock.RLock()
	defer p.lock.RUnlock()
	blob, err := json.Marshal(p.results)
	if err != nil {
		return err
	}
	return ioutil.WriteFile("results.json", blob, 0644)
}

func (p *Parser) Parse(info map[string]int, path string) error {
	defer os.Remove(path)
	out, err := exec.Command("python3", "imports.py", path).Output()
	if err != nil {
		return err
	}
	if len(out) == 0 {
		return nil
	}
	for _, dead := range strings.Split(string(out), ",") {
		info[dead] += 1
	}
	return nil
}

func run() error {
	pkgs := make(chan string)
	var results map[string]map[string]int

	// Load existing results
	if _, err := os.Stat("results.json"); err == nil {
		blob, err := ioutil.ReadFile("results.json")
		if err != nil {
			return err
		}
		if err := json.Unmarshal(blob, &results); err != nil {
			return err
		}
	} else {
		results = map[string]map[string]int{}
	}

	parser := Parser{in: pkgs, results: results}

	// Save results every minute
	ticker := time.NewTicker(60 * time.Second)
	go func() {
		for range ticker.C {
			parser.Save()
		}
	}()

	for i := 0; i < 10; i++ {
		go parser.Work()
	}

	matches, err := filepath.Glob("pypi/web/packages/*/*/*/*")
	if err != nil {
		return err
	}
	for _, pkg := range matches {
		pkgs <- pkg
	}

	close(pkgs)
	parser.wg.Wait()

	return parser.Save()
}
