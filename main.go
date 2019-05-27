package main

import (
	"flag"
	"log"
)

func main() {
	flag.Parse()
	var err error
	switch flag.Arg(0) {
	case "mirror":
		err = mirror()
	case "genpkg":
		err = genpkg()
	case "stats":
		err = stats()
	}
	if err != nil {
		log.Fatal(err)
	}
}
