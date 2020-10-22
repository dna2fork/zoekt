package analysis

import (
	"fmt"
	"bufio"
	"time"
	"os/exec"
	"encoding/json"
	"path/filepath"
)

type CTagsItem struct {
	Type string     `json:"_type"`
	Name string     `json:"name"`
	Path string     `json:"path"`
	// format: /^...$/
	Pattern string  `json:"pattern"`
	Language string `json:"language"`
	Line uint       `json:"line"`
	Kind string     `json:"kind"`
}
type SymbolItem struct {
	Type string `json:"type"`
	Name string `json:"name"`
	Line uint   `json:"line"`
	Kind string `json:"kind"`
}
type SymbolSet struct {
	Language string       `json:"language"`
	Symbols  []SymbolItem `json:"symbols"`
}

func init() {
	var err error
        if CTAGS_BIN == "" {
                CTAGS_BIN, err = exec.LookPath("universal-ctags")
                if err != nil {
                        CTAGS_BIN = ""
                }
        }

        if CTAGS_BIN == "" {
                CTAGS_BIN, err = exec.LookPath("ctags-exuberant")
                if err != nil {
                        CTAGS_BIN = ""
                }
        }
}

func runCtags(filename string, output chan<- *SymbolSet) error {
	if CTAGS_BIN == "" {
		return fmt.Errorf("no ctags found")
	}
	argv := []string{
		"--extras=-fq", "--file-scope=yes", "--excmd=mix",
		"--fields=Klns", "--output-format=json", "--sort=no",
		filename,
	}
	proc := exec.Command(CTAGS_BIN, argv...)
	stdout, err := proc.StdoutPipe()
	if err != nil {
		stdout = nil
	}
	err = proc.Start()
	if err != nil {
		return err
	}
	errChan := make(chan error, 1)
	done := make(chan error, 1)
	go func () {
		err := proc.Wait()
		errChan <- err
	}()
	go func () {
		if stdout == nil {
			return
		}
		var set SymbolSet
		scanner := bufio.NewScanner(stdout)
		scanner.Split(bufio.ScanLines)
		for scanner.Scan() {
			m := scanner.Text()
			var item CTagsItem
			err := json.Unmarshal([]byte(m), &item)
			if err != nil { continue }
			if set.Language == "" { set.Language = item.Language }
			set.Symbols = append(set.Symbols, SymbolItem{item.Type, item.Name, item.Line, item.Kind})
		}
		output <- &set
		done <- nil
	}()
	timeout := time.After(5 * time.Second)
	select {
	case <-timeout:
		proc.Process.Kill()
		return fmt.Errorf("timeout executing ctags")
	case err = <-errChan:
		if err != nil { return err }
	}
	return <-done
}

func GetCtagsSymbols(project IProject, path string) (*SymbolSet, error) {
	out := make(chan *SymbolSet, 1)
	// TODO: check .. in path
	path = filepath.Join(project.GetBaseDir(), path)
	err := runCtags(path, out)
	if err != nil {
		return nil, err
	}
	r := <-out
	return r, nil
}
