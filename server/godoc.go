package server

import (
	"log"
	"net/http"
	"strings"
	"text/template"

	"golang.org/x/tools/godoc"
	"golang.org/x/tools/godoc/static"
	"golang.org/x/tools/godoc/vfs"
	"golang.org/x/tools/godoc/vfs/gatefs"
	"golang.org/x/tools/godoc/vfs/mapfs"
)

type Godoc struct {
	*godoc.Presentation
	fs vfs.NameSpace
}

func NewGodoc(goroot, gopath string) http.Handler {
	vfs.GOROOT = goroot
	fsGate := make(chan bool, 20)
	fs := vfs.NameSpace{}
	fs.Bind("/", gatefs.New(vfs.OS(goroot), fsGate), "/", vfs.BindReplace)
	fs.Bind("/src", gatefs.New(vfs.OS(gopath), fsGate), "/src", vfs.BindAfter)
	fs.Bind("/lib/godoc", mapfs.New(static.Files), "/", vfs.BindReplace)

	corpus := godoc.NewCorpus(fs)
	corpus.Verbose = false
	corpus.MaxResults = 10000
	corpus.IndexEnabled = false
	corpus.IndexFiles = ""
	corpus.IndexDirectory = func(dir string) bool {
		return dir != "/pkg" && !strings.HasPrefix(dir, "/pkg/")
	}
	corpus.IndexThrottle = 0.75
	corpus.IndexInterval = 0
	go func() {
		err := corpus.Init()
		if err != nil {
			log.Fatal(err)
		}
		corpus.RunIndexer()
	}()
	corpus.InitVersionInfo()

	gd := &Godoc{
		fs:           fs,
		Presentation: godoc.NewPresentation(corpus),
	}
	gd.Presentation.ShowTimestamps = false
	gd.Presentation.ShowPlayground = false
	gd.Presentation.DeclLinks = true

	gd.readTemplates()

	return gd
}

func (doc *Godoc) readTemplates() {
	doc.CallGraphHTML = doc.readTemplate("callgraph.html")
	doc.DirlistHTML = doc.readTemplate("dirlist.html")
	doc.ErrorHTML = doc.readTemplate("error.html")
	doc.ExampleHTML = doc.readTemplate("example.html")
	doc.GodocHTML = doc.readTemplate("godoc.html")
	doc.ImplementsHTML = doc.readTemplate("implements.html")
	doc.MethodSetHTML = doc.readTemplate("methodset.html")
	doc.PackageHTML = doc.readTemplate("package.html")
	doc.PackageRootHTML = doc.readTemplate("packageroot.html")
	doc.SearchHTML = doc.readTemplate("search.html")
	doc.SearchDocHTML = doc.readTemplate("searchdoc.html")
	doc.SearchCodeHTML = doc.readTemplate("searchcode.html")
	doc.SearchTxtHTML = doc.readTemplate("searchtxt.html")
}

func (doc *Godoc) readTemplate(name string) *template.Template {
	path := "lib/godoc/" + name

	// use underlying file system fs to read the template file
	// (cannot use template ParseFile functions directly)
	data, err := vfs.ReadFile(doc.fs, path)
	if err != nil {
		log.Fatal("readTemplate: ", err)
	}
	// be explicit with errors (for app engine use)
	t, err := template.New(name).Funcs(doc.FuncMap()).Parse(string(data))
	if err != nil {
		log.Fatal("readTemplate: ", err)
	}
	return t
}
