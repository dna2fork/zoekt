package keyval

import (
	"path/filepath"
	"net/url"
	"os"
	"io/ioutil"
	"strings"
	"errors"
	"fmt"
)

const MAX_SUBKEY_DISPLAY_NUMBER = 4096;

type StorageFilesystem struct{
	Storage
}

type config struct {
	BaseDir string
}

var defaultConfig config

func init() {
	baseDirStr := os.Getenv("KEYVAL_STORAGE_FS_BASE_DIR")
	if baseDirStr == "" {
		return
	}
	baseDir, err := filepath.Abs(baseDirStr)
	fmt.Println("[I] [Storage.Filesystem] baseDir =", baseDir)
	defaultConfig.BaseDir = baseDir
	if err != nil {
		panic("[E] [Storage.Filesystem] cannot assign base directory")
	}
	defaultConfig.BaseDir += string(filepath.Separator)
}

func GetBaseDir() string {
	return defaultConfig.BaseDir
}

func IsKeyvalFSEnabled() bool {
	return defaultConfig.BaseDir != ""
}

func (s *StorageFilesystem) prepare(path string) error {
	path, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	if !strings.HasPrefix(path, defaultConfig.BaseDir) {
		return errors.New("[E] [Storage.Filesystem] hacked file path")
	}
	err = os.MkdirAll(path, 0755)
	if err != nil {
		return err
	}

	return nil
}

func (s *StorageFilesystem) keyUnderLine(part string) string {
	// _ is control character, so we add _ when we meet under line part
	// e.g. _ -> __, ___ -> ____
	if part == "" {
		return part
	}
	for _, ch := range []rune(part) {
		if ch != '_' {
			return part
		}
	}
	return "_" + part
}

func (s *StorageFilesystem) urlUnderLine(part string) string {
	// _ is control character, so we remove _ when we meet under line part
	// e.g. __ -> _, ____ -> ___
	if part == "" {
		return part
	}
	for _, ch := range []rune(part) {
		if ch != '_' {
			return part
		}
	}
	return string([]rune(part)[1:])
}

func (s *StorageFilesystem) UrlToKey(url_ string) string {
	// convert url to key, where the key is a file path
	// url: https://www.github.com/dna2zodiac/keyval
	// key: /baseDir/https/www.github.com/dna2zodiac/keyval/_
	if strings.Index(url_, "://") < 0 {
		if strings.HasPrefix(url_, "//") {
			// e.g. //github.com -> keyval://github.com
			url_ = "keyval:" + url_
		} else {
			// e.g. test -> keyval://test
			url_ = "keyval://" + url_
		}
	}
	parts := strings.Split(url_, "/")
	parts[0] = string([]rune(parts[0])[0:len(parts[0]) - 1])
	for i, part := range parts {
		parts[i] = url.QueryEscape(s.keyUnderLine(part))
	}
	key_ := filepath.Join(parts...)
	key_ = filepath.Join(defaultConfig.BaseDir, key_, "_")
	return key_
}

func (s *StorageFilesystem) KeyToUrl(key_ string) string {
	// convert key to url
	// key: /baseDir/https/www.github.com/dna2zodiac/keyval/_
	// url: https://www.github.com/dna2zodiac/keyval
	if !strings.HasSuffix(key_, "_") {
		return ""
	}
	if !strings.HasPrefix(key_, defaultConfig.BaseDir) {
		return ""
	}
	key_ = string([]rune(key_)[len(defaultConfig.BaseDir):])
	parts := strings.Split(key_, string(filepath.Separator))
	parts[0] = string([]rune(parts[0])[0:len(parts[0]) - 1])
	for i, part := range parts {
		decoded, err := url.QueryUnescape(s.urlUnderLine(part))
		parts[i] = decoded
		if err != nil {
			return ""
		}
	}
	url_ := parts[0] + "://" + strings.Join(parts[1:len(parts) - 1], "/")
	return url_
}

func (s *StorageFilesystem) WithSubkey(key, subkey string) string {
	if filepath.Separator == '/' {
		return filepath.Join(key, subkey)
	}
	return filepath.Join(key, strings.ReplaceAll(subkey, "/", string(filepath.Separator)))
}

func (s *StorageFilesystem) ListSubkey(key string) []string {
	r := make([]string, 0)
	f, err := os.Open(key)
	if err != nil {
		return r
	}
	defer f.Close()
	fileList, err := f.Readdir(MAX_SUBKEY_DISPLAY_NUMBER)
	if err != nil {
		return r
	}
	for _, item := range fileList {
		if item.IsDir() {
			r = append(r, item.Name() + string(filepath.Separator))
		} else {
			r = append(r, item.Name())
		}
	}
	return r
}

func (s *StorageFilesystem) Get(key string) ([]byte, bool) {
	info, err := os.Stat(key)
	if err != nil {
		return nil, false
	}
	m := info.Mode()
	if m & os.ModeDir != 0 || m & os.ModeDevice != 0 {
		return nil, false
	}
	if m & os.ModeNamedPipe != 0 || m & os.ModeSocket != 0 || m & os.ModeIrregular != 0 {
		return nil, false
	}
	r, err := ioutil.ReadFile(key)
	if err != nil {
		return nil, false
	}
	return r, true
}

func (s *StorageFilesystem) Put(key string, value []byte) bool {
	dir, _ := filepath.Split(key)
	err := s.prepare(dir)
	if err != nil {
		return false
	}
	err = ioutil.WriteFile(key, value, 0644)
	if err != nil {
		return false
	}
	return true
}

func (s *StorageFilesystem) Del(key string) bool {
	err := os.Remove(key)
	if err != nil {
		return false
	}
	return true
}

