package server

import (
	"bufio"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/syndtr/goleveldb/leveldb"
)

type Convert struct {
	Goroot  string
	Gopath  string
	Workdir string
	DB      *leveldb.DB
}

func NewConvert(goroot, gopath, dir string) (*Convert, error) {
	db, err := leveldb.OpenFile(filepath.Join(dir, "data"), nil)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	log.Println("new Convert", goroot, gopath)
	return &Convert{
		Goroot:  goroot,
		Gopath:  gopath,
		Workdir: dir,
		DB:      db,
	}, nil
}

func (cvt *Convert) ConvertPackage(pkg string) {
	if Exists(filepath.Join(cvt.Goroot, "src", pkg)) {
		cvt.ConvertDir(filepath.Join(cvt.Goroot, "src", pkg), filepath.Join(cvt.Workdir, "root/src", pkg))
	}
	if Exists(filepath.Join(cvt.Gopath, "src", pkg)) {
		cvt.ConvertDir(filepath.Join(cvt.Gopath, "src", pkg), filepath.Join(cvt.Workdir, "path/src", pkg))
	}
}

func (cvt *Convert) ConvertDir(source, target string) error {
	infos, err := ioutil.ReadDir(source)
	if err != nil {
		return err
	}

	err = os.MkdirAll(target, 0644)
	if err != nil {
		return err
	}

	for _, i := range infos {
		if !i.IsDir() && strings.HasSuffix(i.Name(), ".go") {
			cvt.ConvertFile(filepath.Join(source, i.Name()), filepath.Join(target, i.Name()))
		}
	}
	return nil
}

func (cvt *Convert) ConvertFile(source, target string) error {
	fset := token.NewFileSet()
	os.Remove(target)
	f, err := parser.ParseFile(fset, source, nil, parser.ParseComments)
	if err != nil {
		panic(err)
	}

	for i := range f.Comments {
		cvt.ConvertDoc(f.Comments[i])
	}

	file, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	return format.Node(file, fset, f)
}

func (cvt *Convert) ConvertDoc(doc *ast.CommentGroup) {
	for i := range doc.List {
		val, err := cvt.DB.Get([]byte(doc.List[i].Text), nil)
		if err == nil {
			doc.List[i].Text = string(val)
		}
	}
}

func (cvt *Convert) PutData(key, val string) {
	if key != val {
		cvt.DB.Put([]byte(key), []byte(val), nil)
	}
}

func (cvt *Convert) InputFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		log.Println(err)
		return err
	}
	defer file.Close()

	var (
		ispartcomment bool
		data          string
		datas         []string
	)

	buf := bufio.NewReader(file)
	for {
		s, err := buf.ReadString('\n')
		if err == io.EOF {
			data = data + s
			break
		}
		if err != nil {
			log.Println(err)
			return nil
		}

		if ispartcomment {
			if s == "*/\r\n" {
				ispartcomment = false
				data = strings.Replace(data, "\r\n", "\n", -1) + "*/"
				datas = append(datas, data)
				data = ""
			} else {
				data = data + s

			}
		} else {

			switch {
			case strings.HasPrefix(s, "//"):
				data = data + s
			case s == "\r\n":
				if len(data) > 2 {
					datas = append(datas, data)
				}
				data = ""
			case s == "/*\r\n":
				data = s
				ispartcomment = true
			}
		}
	}
	if len(datas)%2 == 1 {
		datas = append(datas, data)
	}

	for i := 0; i < len(datas); i += 2 {
		if strings.HasPrefix(datas[i], "/*") {
			cvt.PutData(datas[i], datas[i+1])
		} else {
			cvt.InputPart(datas[i], datas[i+1])
		}
	}
	return nil
}

func (cvt *Convert) InputPart(p1, p2 string) {
	str1 := strings.Split(p1, "\r\n")
	str2 := strings.Split(p2, "\r\n")
	length := len(str1)
	if length > len(str2) {
		length = len(str2)
	}
	for i := 0; i < length; i++ {
		cvt.PutData(str1[i], str2[i])
	}
}

func init() {
	// parsefile("/tmp/godoc/cn/root/src/net/http/doc.go")
	// parsefile("/usr/local/go1.13/src/net/http/doc.go")
}

func parsefile(path string) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	// f, err := parser.ParseDir(fset, "/mnt/hgfs/golang/src/github.com/eudore/eudore", nil, parser.ParseComments)
	if err != nil {
		panic(err)
	}
	// fmt.Println([]byte(f.Doc.List[0].Text))
	ast.Print(fset, f)
}

// 判断所给路径文件/文件夹是否存在
func Exists(path string) bool {
	_, err := os.Stat(path) //os.Stat获取文件信息
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

func CopyDir(src, des string) error {
	infos, err := ioutil.ReadDir(src)
	if err != nil {
		return err
	}

	os.MkdirAll(des, 0644)
	for _, i := range infos {
		if i.IsDir() {
			CopyDir(filepath.Join(src, i.Name()), filepath.Join(des, i.Name()))
		} else {
			CopyFile(filepath.Join(src, i.Name()), filepath.Join(des, i.Name()))
		}
	}
	return nil
}

func CopyFile(src, des string) (written int64, err error) {
	srcFile, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer srcFile.Close()

	stat, _ := srcFile.Stat()
	desFile, err := os.OpenFile(des, os.O_CREATE|os.O_WRONLY, stat.Mode())
	if err != nil {
		return 0, err
	}
	defer desFile.Close()

	return io.Copy(desFile, srcFile)
}
