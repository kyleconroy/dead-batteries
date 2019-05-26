# Dead Batteries

[PEP 594](https://www.python.org/dev/peps/pep-0594/) outlines the plan to
deprecate and remove packages from the Python standard library. If accepted in
its current form, PEP 594 will break 3.2% of all published packages on PyPI.

## Results

As of May 25, 2019 there are 5840 packages on PyPI (out of 181,268) that import
packages deprecated by PEP 594.

## Instructions

Before you begin, you'll need Python >= 3.6 and a recent version of Go to run
the code. All commands should be run inside a cloned checkout of the this
repository. First, install
[bandersnatch](https://pypi.org/project/bandersnatch/).

```
python3 -m venv venv
venv/bin/pip install bandersnatch
```

The following command will mirror the latest release of every package on PyPI
to your workstation. You'll need around 500GB of free space on your drive.

```
venv/bin/bandersnatch -c bandersnatch.conf mirror
```

With a complete mirror, you're now ready to process the packages. In a screen
session, run the search program.

```
go run search.go
```

This took a few days to run, as it parses every Python file in each package. I
let this run on an instance on Google Cloud, so I didn't really care that it
was slow. The output is continually saved to `results.json`.

Once the search process is complete, generate the package statistics. 

```
go run results.go
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
