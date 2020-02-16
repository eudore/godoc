package main

import (
	"os"
	"strings"

	"github.com/eudore/godoc/server"
)

// config
var (
	goroot  string
	gopath  string
	workdir = "/tmp/godoc"
	langs   []string
	datas   []string
	addr    = ":8080"
)

func init() {
	for _, arg := range os.Args[1:] {
		pos := strings.IndexByte(arg, '=')
		if pos == -1 {
			continue
		}
		key, val := arg[:pos], arg[pos+1:]
		switch key {
		case "--goroot":
			goroot = val
		case "--gopath":
			gopath = val
		case "--lang":
			langs = append(langs, val)
		case "--data":
			datas = append(datas, val)
		case "--addr":
			addr = val
		}
	}
	if len(langs) == 0 {
		langs = []string{"cn"}
	}
}

func main() {
	srv := server.NewServer(goroot, gopath, workdir, langs)
	go srv.InitData(datas)
	srv.Run(addr)
}
