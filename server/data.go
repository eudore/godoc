package server

import (
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

func (srv *Server) ConvertPackage(lang, pkg string) {
	conv, ok := srv.Convs[lang]
	if ok {
		log.Printf("Server convert package %s %s", lang, pkg)
		conv.ConvertPackage(pkg)
	}
}

func (srv *Server) InputFile(lang, path string) error {
	conv, ok := srv.Convs[lang]
	if ok {
		log.Printf("Server input data file %s %s", lang, path)
		return conv.InputFile(path)
	}
	return nil
}

func (srv *Server) InitData(paths []string) {
	for _, path := range paths {
		infos, err := ioutil.ReadDir(path)
		if err != nil {
			log.Println(err)
			return
		}
		for _, info := range infos {
			if !info.IsDir() {
				pkg, lang := parseName(info.Name())
				if lang != "" {
					srv.InputFile(lang, filepath.Join(path, info.Name()))
					srv.ConvertPackage(lang, pkg)
				}
			}
		}
		go updateGit(path)
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Printf("info fsnotify error: %v\n", err)
		return
	}
	for _, path := range paths {
		log.Printf("fsnotify add watch %s\n", path)
		watcher.Add(path)
	}
	defer log.Println("end")
	defer watcher.Close()
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				break
			}

			name := filepath.Base(event.Name)
			if name[0] != '.' && event.Op&fsnotify.Write == fsnotify.Write {
				pkg, lang := parseName(name)
				log.Println("fsnotify write file: ", event.Name)
				srv.InputFile(lang, event.Name)
				srv.ConvertPackage(lang, pkg)
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				break
			}
			log.Println("notify watcher error:", err)
		}
	}
}

func parseName(name string) (string, string) {
	strs := strings.Split(name, ".")
	if len(strs) < 3 {
		return "", ""
	}
	return strings.Replace(strs[0], "-", "/", -1), strs[len(strs)-2]
}

func updateGit(path string) {
	if !Exists(filepath.Join(path, ".git")) {
		return
	}
	for {
		time.Sleep(60 * time.Second)
		cmd := exec.Command("/usr/bin/git", "pull")
		cmd.Dir = path
		cmd.Env = os.Environ()
		err := cmd.Run()
		if err != nil {
			log.Printf("git pull %s error: %v\n", path, err)
			return
		}
	}
}
