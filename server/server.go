package server

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
)

type Server struct {
	Goroot  string
	Gopath  string
	Workdir string
	Godoc   http.Handler
	Langs   map[string]http.Handler
	Convs   map[string]*Convert
}

func NewServer(goroot, gopath, workdir string, langs []string) *Server {
	log.Println("goroot:", goroot)
	log.Println("gopath:", gopath)
	log.Println("workdir:", workdir)
	log.Println("langs:", langs)
	srv := &Server{
		Goroot:  goroot,
		Gopath:  gopath,
		Workdir: workdir,
		Godoc:   NewGodoc(goroot, gopath),
		Langs:   make(map[string]http.Handler),
		Convs:   make(map[string]*Convert),
	}
	log.Println("create server success")

	var err error
	for _, lang := range langs {
		log.Println("init lang", lang)
		path := filepath.Join(workdir, lang)
		os.MkdirAll(path+"/root/src", 0644)
		os.MkdirAll(path+"/path/src", 0644)
		if !Exists(filepath.Join(goroot, "/root/doc")) {
			err = CopyDir(filepath.Join(goroot, "doc"), filepath.Join(path, "/root/doc"))
			log.Printf("init create lang %s doc error %v", lang, err)
		}
		srv.Langs[lang] = NewGodoc(path+"/root", path+"/path")
		srv.Convs[lang], err = NewConvert(srv.Goroot, srv.Gopath, filepath.Join(srv.Workdir, lang))
		if err != nil {
			log.Println("init lang error", err)
			panic(err)
		}
	}
	log.Println("init success")
	return srv
}

func (srv *Server) Run(addr string) {
	mux := http.NewServeMux()
	for lang, h := range srv.Langs {
		length := len(lang) + 6
		handle := h
		mux.HandleFunc("/lang/"+lang+"/", func(w http.ResponseWriter, r *http.Request) {
			r.URL.Path = r.URL.Path[length:]
			handle.ServeHTTP(w, r)
			w.Write([]byte(`<script src="/lib/godoc/init.js" defer=""></script>`))
		})
	}
	mux.HandleFunc("/lib/godoc/init.js", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/javascript")
		w.Write([]byte(`
			var lang = "/lang/cn"
			for(var i of document.getElementsByTagName('a')) {
				if (i.href.indexOf(location.origin)!=-1&&i.href.indexOf(location.origin+lang)==-1) {
					i.href=i.href.replace(location.origin,location.origin+lang)
				}
			}	`))
	})
	mux.HandleFunc("/lang/data", srv.putdata)
	mux.Handle("/", srv.Godoc)
	s := http.Server{
		Addr:    addr,
		Handler: mux,
	}
	log.Println("start server", addr)
	err := s.ListenAndServe()
	log.Println(err)
}

func (srv *Server) putdata(w http.ResponseWriter, r *http.Request) {

}
