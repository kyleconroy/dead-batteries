package main

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type Parser struct {
	client  *http.Client
	lock    sync.RWMutex
	wg      sync.WaitGroup
	in      chan string
	results map[string]result
}

func (p *Parser) Work() {
	p.wg.Add(1)
	defer p.wg.Done()
	for {
		select {
		case path, ok := <-p.in:
			if path == "" {
				return
			}
			if err := p.Unpack(path); err != nil {
				log.Printf("result=failure package=%s err=%s\n", filepath.Base(path), err)
			} else {
				log.Printf("result=success package=%s\n", filepath.Base(path))
			}
			if !ok {
				return
			}
		}
	}
}

func (p *Parser) Unpack(pkg string) error {
	p.lock.RLock()
	_, ok := p.results[pkg]
	p.lock.RUnlock()
	if ok {
		return nil
	}

	// Read metadata file
	filename := filepath.Join("meta", pkg+".json")
	if _, err := os.Stat(filename); err != nil {
		return err
	}
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	dec := json.NewDecoder(f)
	var pypi PyPI
	if err := dec.Decode(&pypi); err != nil {
		return err
	}

	f.Close()

	// Get URL to download
	pkgurl, version, err := pypi.LatestSource()
	if err != nil {
		return err
	}

	// Download file
	archive, err := ioutil.TempFile("", "pep594")
	if err != nil {
		return err
	}
	defer os.Remove(archive.Name())
	defer archive.Close()

	// Download the package
	resp, err := p.client.Get(pkgurl)
	if err != nil {
		return fmt.Errorf("%s: %s", pkgurl, err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("expected 200, not %d", resp.StatusCode)
	}
	defer resp.Body.Close()
	_, err = io.Copy(archive, resp.Body)
	if err != nil {
		return err
	}

	// Explicitly close files
	archive.Close()
	resp.Body.Close()

	results := map[string]int{}
	errors := map[string]int{}

	if strings.HasSuffix(pkgurl, ".tar.gz") {
		r, err := os.Open(archive.Name())
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
			if err := p.Parse(results, errors, tmpfile.Name()); err != nil {
				return err
			}
		}
	} else if strings.HasSuffix(pkgurl, ".whl") || strings.HasSuffix(pkgurl, ".zip") || strings.HasSuffix(pkgurl, ".egg") {
		r, err := zip.OpenReader(archive.Name())
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
			if err := p.Parse(results, errors, tmpfile.Name()); err != nil {
				return err
			}
		}
	} else {
		return fmt.Errorf("unknown format: %s", pkgurl)
	}

	p.lock.Lock()
	p.results[pkg] = result{
		URL:     pkgurl,
		Version: version,
		Imports: results,
		Errors:  errors,
	}
	p.lock.Unlock()

	return nil
}

func (p *Parser) Save() error {
	p.lock.RLock()
	defer p.lock.RUnlock()
	blob, err := json.MarshalIndent(p.results, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile("results.json", blob, 0644)
}

type parsereq struct {
	Source string `json:"path"`
}

type parseresp struct {
	Imports []string `json:"imports"`
	Error   string   `json:"error"`
}

func (p *Parser) Parse(info, errors map[string]int, path string) error {
	defer os.Remove(path)

	blob, err := json.Marshal(&parsereq{Source: path})
	if err != nil {
		return err
	}

	// TODO: Talk to a server over HTTP instead
	resp, err := http.Post("http://127.0.0.1:4000", "application/json", bytes.NewBuffer(blob))
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s <%d>", path, resp.StatusCode)
	}
	var pyresp parseresp
	defer resp.Body.Close()
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&pyresp); err != nil {
		return err
	}
	if pyresp.Error != "" {
		errors[pyresp.Error] += 1
	}
	for _, dead := range pyresp.Imports {
		info[dead] += 1
	}
	return nil
}

type result struct {
	URL     string         `json:"url"`
	Version string         `json:"version"`
	Imports map[string]int `json:"imports,omitempty"`
	Errors  map[string]int `json:"errors,omitempty"`
}

func scan() error {
	pkgs := make(chan string)
	var results map[string]result

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
		results = map[string]result{}
	}

	parser := Parser{in: pkgs, results: results, client: &http.Client{}}

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

	// Load Python3 packages
	var packages []string
	blob, err := ioutil.ReadFile("python3-packages.json")
	if err != nil {
		return err
	}
	if err := json.Unmarshal(blob, &packages); err != nil {
		return err
	}

	// pkgs <- packages[0]
	for _, p := range packages {
		pkgs <- p
	}
	close(pkgs)

	parser.wg.Wait()
	return parser.Save()
}
