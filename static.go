package larixmid

import (
	"net/http"
	"strings"
	"path"
	"os"
)

type Static struct {
	Dir http.FileSystem
	Prefix string
	IndexFile string
}

func (s *Static) ServeHTTP(w http.ResponseWriter, r * http.Request, next http.HandlerFunc) {
	if r.Method != "GET" && r.Method != "Head" {
		next(w, r)
		return
	}

	file := r.URL.Path
	if s.Prefix != "" {
		if !strings.HasPrefix(file, s.Prefix) {
			next(w, r)
			return
		}

		file = file[len(s.Prefix):]
		if file != "" && file[0] != '/' {
			next(w, r)
			return
		}
	}

	f, err := s.Dir.Open(file)
	if err != nil {
		next(w, r)
		return
	}
	defer f.Close()

	fstat, err := f.Stat()
	if err != nil {
		next(w, r)
		return
	}

	// try to use index
	if fstat.IsDir() {
		// missing trailing /
		if !strings.HasSuffix(r.URL.Path, "/") {
			http.Redirect(w, r, r.URL.Path+"/", http.StatusFound)
			return
		}

		file = path.Join(file, s.IndexFile)
		f, err = os.Open(file)
		if err != nil {
			next(w, r)
			return
		}
		defer f.Close()

		fstat, err = f.Stat()
		if err != nil {
			next(w, r)
			return
		}
	}

	http.ServeContent(w, r, file, fstat.ModTime(), f)
}

func NewStatic(dir http.FileSystem) *Static {
	return &Static{
		Dir: dir,
		Prefix: "",
		IndexFile: "index.html",
	}
}
