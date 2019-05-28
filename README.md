# Dead Batteries

venv/bin/pip install flask
venv/bin/pip install gunicorn
venv/bin/pip install gunicorn[gevent]
venv/bin/gunicorn -k gevent -w 20 -b 127.0.0.1:4000 imports:app

[PEP 594](https://www.python.org/dev/peps/pep-0594/) outlines the plan to
deprecate and remove packages from the Python standard library. If accepted in
its current form, PEP 594 will break 3.2% of all published packages on PyPI.

## Results

As of May 25, 2019 there are 5840 packages on PyPI (out of 181,268) that import
packages deprecated by PEP 594.

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

Some packages listed in simple.html do not have any files on PyPI
TODO: Provide an example

* `simple.html`: contains a list of all known packages on PyPI  
* `meta/*.json`: contains package metadata for each package
* `python3-packages.json`: contains a list of packages that support Python3

With a complete mirror, you're now ready to process the packages. In a screen
session, run the search program.

TODO: Add current stats
94680 packages

```
./dead-battery scan-imports
```

This took a few days to run, as it parses every Python file in each package. I
let this run on an instance on Google Cloud, so I didn't really care that it
was slow. The output is continually saved to `results.json`.

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
5840
```
