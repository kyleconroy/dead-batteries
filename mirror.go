package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/sync/errgroup"
)

var errNotFound = errors.New("not found")

func download(uri, filename string) error {
	resp, err := http.Get(uri)
	if err != nil {
		return fmt.Errorf("%s: %s", uri, err)
	}
	if resp.StatusCode == http.StatusNotFound {
		return errNotFound
	}
	defer resp.Body.Close()
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, resp.Body)
	if err != nil {
		return err
	}
	return nil
}

func downloadIndex() error {
	// Download all packages on PyPI
	if _, err := os.Stat("simple.html"); err != nil {
		return download("https://pypi.org/simple/", "simple.html")
	}
	return nil
}

func packageNames() ([]string, error) {
	packages := []string{}

	// Parse package names
	r, err := os.Open("simple.html")
	if err != nil {
		return packages, err
	}
	doc, err := html.Parse(r)
	if err != nil {
		return packages, err
	}

	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, a := range n.Attr {
				if a.Key == "href" {
					parts := strings.Split(a.Val, "/")
					packages = append(packages, parts[2])
					break
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)
	return packages, nil
}

type Release struct {
	PythonVersion string `json:"python_version"`
}

type PackageInfo struct {
	Classifiers []string `json:"classifiers"`
}

type PyPI struct {
	Info     PackageInfo          `json:"info"`
	Versions map[string][]Release `json:"releases"`
}

func (p PyPI) SupportsPython3() bool {
	for _, c := range p.Info.Classifiers {
		if strings.HasPrefix(c, "Programming Language :: Python :: 3") {
			return true
		}
	}
	for _, releases := range p.Versions {
		for _, release := range releases {
			if strings.Contains(release.PythonVersion, "/") {
				continue
			}
			switch release.PythonVersion {
			// Python 2
			case "py2":
			case "py22":
			case "py23":
			case "py24":
			case "py25":
			case "py26":
			case "py27":
			case "py2.7":
			case "2":
			case "2.2":
			case "2.3":
			case "2.4":
			case "2.5":
			case "2.6":
			case "2.7":
			case "2.7.6":
			case "cp2":
			case "cp22":
			case "cp23":
			case "cp24":
			case "cp25":
			case "cp26":
			case "cp27":
			case "cpy27":
			case "pp2510":
			case "pp260":
			case "pp27":
			case "pp271":

			// Python 3
			case "py3":
				return true
			case "py32":
				return true
			case "py33":
				return true
			case "py34":
				return true
			case "py35":
				return true
			case "py36":
				return true
			case "py37":
				return true
			case "cp31":
				return true
			case "cp32":
				return true
			case "cp33":
				return true
			case "cp34":
				return true
			case "cp35":
				return true
			case "cp36":
				return true
			case "cp37":
				return true
			case "cp34.cp35.cp36":
				return true
			case "cp34.cp35.cp36,cp37":
				return true
			case "cp34.cp35.cp36.cp37":
				return true
			case "cp35.cp36.cp37":
				return true
			case "cp35.cp36.cp37.cp38":
				return true
			case "cp36.cp37":
				return true
			case "3":
				return true
			case "3.0":
				return true
			case "3.1":
				return true
			case "3.2":
				return true
			case "3.3":
				return true
			case "3.4":
				return true
			case "3.5":
				return true
			case "3.5.1":
				return true
			case "3.6":
				return true
			case "3.7":
				return true
			case "pp3510":
				return true
			case "pp360":
				return true
			case "py36.py35":
				return true
			case "py34+":
				return true

			// Either
			case "any":
				return true
			case "py2.py3":
				return true
			case "py3.py2":
				return true
			case "py27,py36,py37":
				return true
			case "py27.py32.py33":
				return true
			case "py27.py3":
				return true
			case "py2.py30":
				return true
			case "py27.py36.py37":
				return true

			// Unknown
			case "source":
			case "":
			case "smart":
			case "software":
			default:
				log.Printf("unknown version: %s", release.PythonVersion)
				// panic(release.PythonVersion)
			}
		}
	}
	return false
}

func downloadMetadata() error {
	packages, err := packageNames()
	if err != nil {
		return err
	}
	if err := os.MkdirAll("meta", 0644); err != nil {
		return err
	}

	pkgchan := make(chan string)
	var g errgroup.Group
	for i := 0; i <= 100; i += 1 {
		g.Go(func() error {
			for p := range pkgchan {
				filename := filepath.Join("meta", p+".json")
				if _, err := os.Stat(filename); err == nil {
					continue
				}
				log.Printf("fetch package metadata: %s", p)
				uri := fmt.Sprintf("https://pypi.python.org/pypi/%s/json", p)
				err := download(uri, filename)
				if err == errNotFound {
					continue
				}
				if err != nil {
					return err
				}
			}
			return nil
		})
	}

	for _, p := range packages {
		pkgchan <- p
	}
	close(pkgchan)

	return g.Wait()
}

func filterPackages() error {
	packages, err := packageNames()
	if err != nil {
		return err
	}

	py3Packages := []string{}
	for _, p := range packages {
		filename := filepath.Join("meta", p+".json")
		if _, err := os.Stat(filename); err != nil {
			continue
		}
		f, err := os.Open(filename)
		if err != nil {
			return err
		}
		dec := json.NewDecoder(f)
		var pypi PyPI
		if err := dec.Decode(&pypi); err != nil {
			f.Close()
			return err
		}
		if pypi.SupportsPython3() {
			py3Packages = append(py3Packages, p)
		}
		f.Close()
	}

	// Parse package names
	blob, err := json.Marshal(py3Packages)
	if err != nil {
		return err
	}

	return ioutil.WriteFile("python3-packages.json", blob, 0644)
}

func redo() error {
	if err := downloadIndex(); err != nil {
		return err
	}
	if err := downloadMetadata(); err != nil {
		return err
	}
	if err := filterPackages(); err != nil {
		return err
	}
	return nil
}
