# Dead Batteries

[PEP 594](https://www.python.org/dev/peps/pep-0594/) outlines the plan to
deprecate and remove packages from the Python standard library. If accepted in
its current form, PEP 594 will break 3.8% of all Python 3 packages on PyPI.

## Results

As of May 2019 there are 3604 Python 3 packages on PyPI (out of 94,680) that
import packages deprecated by PEP 594.

```
  total packages: 181225
  valid packages: 178642
python3 packages: 94680
scanned packages: 92779
 broken packages: 3604
```

## Instructions

Before you begin, you'll need Python >= 3.6 and a recent version of Go to run
the code. All commands should be run inside a cloned checkout of the this
repository. First, build the binary.

```
go build
```

Next, you'll download the metadata for every package on PyPI, which will take
up about ~500MB of space.

```
./dead-battery mirror
```

* `simple.html`: contains a list of all known packages on PyPI  
* `meta/*.json`: contains package metadata for each package
* `python3-packages.json`: contains a list of packages that support Python3

With a complete mirror, you're now ready to scan packages for imports of
deprecated packages. Open up a new shell to install and run the Python service.

```
python3 -m venv venv
venv/bin/pip install flask
venv/bin/pip install gunicorn
venv/bin/pip install gunicorn[gevent]
venv/bin/gunicorn -k gevent -w 10 -b 127.0.0.1:4000 imports:app
```

The service exposes a HTTP/JSON interface to the `ast` package. It parses a
given file and returns any deprecated imports as well as parsing errors. 

```
// INPUT
{
  "path": "/path/to/python/file"
}

// OUTPUT
{ 
  "imports": {
    "imp": 2
  },
  "errors": {
    "syntax-error": 2
  }
}
```

With that running, you can now start the scan. On my laptop, the scan took
about an hour to complete. The output is continually saved to `results.json`.

```
./dead-battery scan
```

Once the search process is complete, generate the package statistics. 

```
./dead-battery stats
```

Two new files have been created: `import.csv` and `packages.json`. The CSV file
contains the total number of imports for each deprecated standard library
package. The JSON file contains a list of every package that imports one of
these deprecated packages, along with a link to the package on PyPI.

A quick jump into the Python interpreter gives us the total number of packages
affected by PEP 594.

```
>>> import json
>>> len(json.load(open('packages.json')))
3604
```

## Methodology 

Since PEP 594 only affects Python 3, I needed to filter out packages that don't
support Python 3. For each package, I first looked for any classifiers with the
prefix `Programming Language :: Python :: 3`. Next, I checked the
`python_version` of the latest release. It's a bit messy, but the code to do so
can be found [here][mirror].

```json
{
    "info": {
        "classifiers": [
            "Programming Language :: Python :: 2.7",
            "Programming Language :: Python :: 3"
        ],
        "version": "2.22.0"
    },
    "releases": {
        "2.22.0": [
            {
                "packagetype": "bdist_wheel",
                "python_version": "py2.py3",
                "url": "https://files.pythonhosted.org/.../requests-2.22.0-py2.py3-none-any.whl"
            }
        ]
    }
}
```

If you know a better way to check for Python 3 compatibility, please reach out
or open an issue.

[mirror]: https://github.com/kyleconroy/dead-batteries/blob/master/mirror.go#L114
