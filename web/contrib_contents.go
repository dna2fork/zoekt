package web

import (
	"net/http"
	"path/filepath"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"
	"log"
	"bytes"

	"github.com/google/zoekt/contrib/analysis"
)

func escapeQuery(queryStr string) string {
	// ( <-> \(; ) <-> \); \) -> ); [ <-> \[; ] <-> \]
	trQueryStr := regexStringEscape(queryStr, "(")
	trQueryStr = regexStringEscape(trQueryStr, ")")
	trQueryStr = regexStringEscape(trQueryStr, "[")
	trQueryStr = regexStringEscape(trQueryStr, "]")
	return trQueryStr
}

func regexStringEscape(origin string, ch string) string {
	// used in 
	parts := strings.Split(origin, ch)
	N := len(parts) - 1
	for i := 0; i < N; i ++ {
		part := parts[i]
		L := len(part)
		C := 0
		for j := L-1; j >= 0; j-- {
			if part[j] != '\\' {
				break
			}
			C ++
		}
		if C % 2 == 0 {
			parts[i] = fmt.Sprintf("%s\\%s", parts[i], ch)
		} else {
			runes := []rune(part)
			parts[i] = fmt.Sprintf("%s%s", string(runes[0:L-1]), ch)
		}
	}
	return strings.Join(parts, "")
}

func jsonText (json string) string {
	json = strings.Replace(json, "\\", "\\\\", -1)
	json = strings.Replace(json, "\n", "\\n", -1)
	json = strings.Replace(json, "\r", "\\r", -1)
	json = strings.Replace(json, "\t", "\\t", -1)
	json = strings.Replace(json, "\"", "\\\"", -1)
	return json
}

type ServerAuthBasic struct {
	FileName string
	Value string
	Mtime time.Time
	Watcher *time.Timer
}

func watchAuthBasic(t *time.Timer, d time.Duration, a *ServerAuthBasic) {
	<- t.C
	if (a.checkModified()) {
		a.loadBasicAuth()
	}
	t.Reset(d)
	go watchAuthBasic(t, d, a)
}

func (a *ServerAuthBasic) checkAuth(r *http.Request) bool {
	if a.FileName == "" {
		return true
	}
	if a.Value == "" {
		a.loadBasicAuth()
	}
	if a.Watcher == nil {
		d := time.Minute
		a.Watcher = time.NewTimer(d)
		go watchAuthBasic(a.Watcher, d, a)
	}

	value := strings.Trim(r.Header.Get("Authorization"), " \r\n\t")
	if value == a.Value {
		return true
	}
	return false
}

func (a *ServerAuthBasic) loadBasicAuth() {
	file, err := os.Open(a.FileName)
	if err != nil {
		log.Printf("failed to load basic auth: %v", err)
		return
	}
	defer file.Close()
	buf, err := ioutil.ReadAll(file)
	if err != nil {
		log.Printf("failed to load basic auth: %v", err)
		return
	}
	nextValue := strings.Trim(string(buf), " \r\n\t")
	if (a.Value == "" && nextValue == "" && a.Watcher == nil) || (a.Value != "" && nextValue == "") {
		log.Printf("set [empty] value to basic auth ...")
	}
	a.Value = nextValue

	stat, err := file.Stat()
	if err == nil {
		a.Mtime = stat.ModTime()
	}
}

func (a *ServerAuthBasic) checkModified() bool {
	if a.FileName == "" {
		return false
	}
	file, err := os.Open(a.FileName)
	if err != nil {
		return false
	}
	stat, err := file.Stat()
	if err != nil {
		return false
	}
	if a.Mtime.Equal(stat.ModTime()) {
		return false
	}
	return true
}

func (s *Server) serveFSPrint(w http.ResponseWriter, r *http.Request) {
	if !s.BasicAuth.checkAuth(r) {
		w.WriteHeader(401)
		w.Write(bytes.NewBufferString("Not authenticated.").Bytes())
		return
	}
	err := s.serveFSPrintErr(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusTeapot)
	}
}

func (s *Server) serveFSPrintErr(w http.ResponseWriter, r *http.Request) error {
	qvals := r.URL.Query()
	fileStr := qvals.Get("f")
	repoStr := qvals.Get("r")
	// var buf bytes.Buffer
	path := fmt.Sprintf("%s/%s%s", s.SourceBaseDir, repoStr, fileStr)
	if !validatePath(path) {
		w.Write([]byte(`{"error":403, "reason": "hacking detcted"}`))
		return nil
	}
	result := isDirectory(path)
	if result == 1 {
		return sendDirectoryContents(w, path)
	} else if result == 0 {
		return sendFileContents(w, path)
	} // else r == -1: err / not exists
	return nil
}

func combileOneItemDirectory(dirname string, basename string) string {
	path := filepath.Join(dirname, basename)
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return basename
	}
	if len(files) != 1 {
		return basename
	}
	if isDirectory(filepath.Join(path, files[0].Name())) != 1 {
		return basename
	}
	return combileOneItemDirectory(dirname, filepath.Join(basename, files[0].Name()))
}

func sendDirectoryContents(w http.ResponseWriter, path string) error {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		w.Write([]byte(`{"error":500}`))
		return err
	}
	buf := `{"directory":true, "contents":[`
	item_tpl := `{"name":"%s"},`
	for _, file := range files {
		name := file.Name()
		subpath := filepath.Join(path, name)
		if isDirectory(subpath) == 1 {
			name = fmt.Sprintf("%s/", combileOneItemDirectory(path, name))
		}
		buf = fmt.Sprintf(`%s%s`, buf, fmt.Sprintf(item_tpl, jsonText(name)))
	}
	buf = fmt.Sprintf(`%snull]}`, buf)
	w.Write([]byte(buf))
	return nil
}

func isBinary(data []byte, n int) bool {
	for index, ch := range ([]rune(string(data[0:n]))) {
		if index >= n {
			break
		}
		if ch == '\x00' {
			return true
		}
	}
	return false
}

func sendFileContents(w http.ResponseWriter, path string) error {
	// TODO: if file is too large, return error
	file, err := os.Open(path)
	if err != nil {
		w.Write([]byte(`{"error":500}`))
		return err
	}
	defer file.Close()
	buf := make([]byte, 4096)
	n, err := file.Read(buf)
	if err != nil {
		w.Write([]byte(`{"error":500}`))
		return err
	}
	if n > 0 {
		if isBinary(buf, n) {
			w.Write([]byte(`{"error":403, "reason":"binary file"}`))
			return nil
		}
		_, err = file.Seek(0, 0)
		if err != nil {
			w.Write([]byte(`{"error":500}`))
			return err
		}
		buf, err = ioutil.ReadAll(file)
		if err != nil {
			w.Write([]byte(`{"error":500}`))
			return err
		}
		w.Write([]byte( fmt.Sprintf(`{"file":true, "contents":"%s"}`, jsonText(string(buf))) ))
		return nil
	}
	w.Write([]byte(`{"file":true, "contents":""}`))
	return nil
}

func validatePath(path string) bool {
	parts := strings.Split(path, "/")
	for _, part := range parts {
		if (part == "..") {
			return false
		}
	}
	return true
}

func checkFileExists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false;
	}
	return true;
}

func isDirectory(path string) int {
	f, err := os.Stat(path);
	if err != nil {
		return -1;
	}
	if (f.Mode() & os.ModeSymlink != 0) {
		linked, err := os.Readlink(path)
		if err != nil {
			return -1;
		}
		f, err = os.Stat(linked)
		if err != nil {
			return -1;
		}
	}
	if (f.Mode().IsDir()) {
		return 1;
	}
	return 0;
}

func (s *Server) serveScmPrint(w http.ResponseWriter, r *http.Request) {
	if !s.BasicAuth.checkAuth(r) {
		w.WriteHeader(401)
		w.Write(bytes.NewBufferString("Not authenticated.").Bytes())
		return
	}

	if analysis.P4_BIN == "" || analysis.GIT_BIN == "" {
		w.Write([]byte(`{"error":403, "reason": "git/p4 not found"}`))
		return
	}

	qvals    := r.URL.Query()
	action   := qvals.Get("a")
	fileStr  := qvals.Get("f")
	repoStr  := qvals.Get("r")
	revision := qvals.Get("x")

	baseDir  := fmt.Sprintf("%s/%s", s.SourceBaseDir, repoStr)
	project  := analysis.NewProject(repoStr, baseDir)
	if project == nil {
		w.Write([]byte(fmt.Sprintf(`{"error":403, "reason": "'%s' not supported nor found"}`, repoStr)))
		return
	}
	path := fmt.Sprintf("%s/%s%s", s.SourceBaseDir, repoStr, fileStr)

	log.Println(action, "\t", path, revision);
	if !validatePath(path) {
		w.Write([]byte(`{"error":403, "reason": "hacking detcted"}`))
		return
	}

	switch action {
	case "get":
		if strings.HasSuffix(fileStr, "/") {
			fileList4aGet, err := project.GetDirContents(fileStr, revision)
			if err != nil {
				w.Write([]byte( fmt.Sprintf(`{"error":403, "reason": "%s"}`, jsonText(err.Error())) ))
				return
			}
			sendScmDirectoryContents(w, fileList4aGet)
		} else {
			fileBin4aGet, err := project.GetFileBinaryContents(fileStr, revision)
			if err != nil {
				w.Write([]byte( fmt.Sprintf(`{"error":403, "reason": "%s"}`, jsonText(err.Error())) ))
				return
			}
			sendScmFileContents(w, fileBin4aGet)
		}
	default:
		w.Write([]byte( fmt.Sprintf(`{"error":403, "reason": "'%s' not support"}`, jsonText(action)) ))
	}
}

func sendScmFileContents(w http.ResponseWriter, buf []byte) {
	n := len(buf)
	if n > 4096 { n = 4096 }
	if isBinary(buf, n) {
		w.Write([]byte(`{"error":403, "reason":"binary file"}`))
		return
	}
	w.Write([]byte( fmt.Sprintf(`{"file":true, "contents":"%s"}`, jsonText(string(buf))) ))
}

func sendScmDirectoryContents(w http.ResponseWriter, nameList []string) {
	buf := `{"directory":true, "contents":[`
	item_tpl := `{"name":"%s"},`
	for _, name := range nameList {
		buf += fmt.Sprintf(item_tpl, jsonText(name))
	}
	buf += "null}"
	w.Write([]byte(buf))
	return
}
