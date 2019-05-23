package main

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

type Parser struct {
	wg sync.WaitGroup
	in chan string
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
			p.Parse(path)
		}
	}
}

func (p *Parser) Parse(path string) {
	defer os.Remove(path)
	fmt.Println(path)
	out, err := exec.Command("python3", "imports.py", path).Output()
	if err != nil {
		return
	}
	if len(out) == 0 {
		return
	}
	fmt.Println(strings.Split(string(out), ","))
}

func run() error {
	files := make(chan string)
	parser := Parser{in: files}

	for i := 0; i < 10; i++ {
		go parser.Work()
	}

	matches, err := filepath.Glob("pypi/web/packages/*/*/*/*")
	if err != nil {
		return err
	}
	for _, pkg := range matches {
		if err := unpack(pkg, files); err != nil {
			return err
		}
	}

	close(files)
	parser.wg.Wait()

	return nil
}

func unpack(pkg string, files chan string) error {
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
			files <- tmpfile.Name()
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
			files <- tmpfile.Name()
		}
	} else {
		fmt.Println("unknown", pkg)
	}
	return nil
}
