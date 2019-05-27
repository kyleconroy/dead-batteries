package main

import "testing"

func TestTrimFormat(t *testing.T) {
	tests := map[string]string{
		"2to3-1.0-py3-none-any.whl":        "2to3-1.0",
		"zymbit-trequests-0.9.5.tar.gz":    "zymbit-trequests-0.9.5",
		"zyklus-0.2-py2.7.egg":             "zyklus-0.2",
		"zzhmodule-1.4.0.zip":              "zzhmodule-1.4.0",
		"AIS.py-0.2.2.linux-x86_64.tar.gz": "AIS.py-0.2.2.linux-x86_64",
	}
	for archive, expected := range tests {
		if trimFormat(archive) != expected {
			t.Errorf("expected %s; got %s", expected, trimFormat(archive))
		}
	}
}

type versionTest struct {
	input string
	pkg   string
	ver   string
}

func TestParseVersion(t *testing.T) {
	tests := []versionTest{
		{"AIS.py-0.2.2.linux-x86_64", "AIS.py", "0.2.2"},
		{"M2CryptoWin32-0.21.1-3", "M2CryptoWin32", "0.21.1-3"},
		{"FelloWiki-0.01a1.dev-r36", "FelloWiki", "0.01a1.dev-r36"},
		{"js.json2-2011-02-23", "js.json2", "2011-02-23"},
		{"hgforest-crew-dev", "hgforest-crew", "dev"},
		{"django-xe-currencies", "django-xe-currencies", ""},
	}
	for _, test := range tests {
		pkg, ver := split(test.input)
		if pkg != test.pkg {
			t.Errorf("%s: expected package %s; got %s", test.input, test.pkg, pkg)
		}
		if ver != test.ver {
			t.Errorf("%s: expected version %s; got %s", test.input, test.ver, ver)
		}
	}
}
