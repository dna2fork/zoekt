package analysis

import (
	"os/exec"
	"sync"
	"os"
	"io"
	"bufio"
	"strings"
)

type execOutputProcessor func (proc *exec.Cmd, stdout, stderr io.ReadCloser) error
type execLinesProcessor func (line string)
type execBytesProcessor func (stdout io.ReadCloser)

func Exec2Lines(cmd string, fn execLinesProcessor) error {
	return doExec(cmd, func (proc *exec.Cmd, stdout, stderr io.ReadCloser) error {
		if err := proc.Start(); err != nil {
			return err
		}
		listener := &sync.WaitGroup{}
		listener.Add(2)
		go watchTextOutput(proc, listener, stdout, fn)
		go watchTextOutput(proc, listener, stderr, fn)
		listener.Wait()
		proc.Wait()
		return nil
	})
}

func watchTextOutput(proc *exec.Cmd, listener *sync.WaitGroup, stream io.ReadCloser, fn execLinesProcessor) {
	defer listener.Done()
	if fn == nil {
		return
	}
	scanner := bufio.NewScanner(stream)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		m := scanner.Text()
		fn(m)
	}
}

func Exec2Bytes(cmd string, fn execBytesProcessor) error {
	return doExec(cmd, func (proc *exec.Cmd, stdout, _stderr io.ReadCloser) error {
		if err := proc.Start(); err != nil {
			return err
		}
		listener := &sync.WaitGroup{}
		listener.Add(1)
		go watchByteOutput(proc, listener, stdout, fn)
		listener.Wait()
		proc.Wait()
		return nil
	})
}

func watchByteOutput(proc *exec.Cmd, listener *sync.WaitGroup, stream io.ReadCloser, fn execBytesProcessor) {
	defer listener.Done()
	if fn == nil {
		return
	}
	fn(stream)
}

func doExec(cmd string, fn execOutputProcessor) error {
	// TEST=1 AND=2 ls -a -l
	// ^      ^     ^  ^--^-- args
	// |      |      \--> cmd
	// \------\-----> env
	argv := strings.Fields(cmd)
	cmdIndex := 0
	for i, value := range argv {
		if !strings.Contains(value, "=") {
			break
		}
		cmdIndex = i + 1
	}
	bin := argv[cmdIndex]
	proc := exec.Command(bin, argv[cmdIndex+1:]...)
	proc.Env = append(os.Environ(), argv[0:cmdIndex]...)
	stdout, err := proc.StdoutPipe()
	if err != nil {
		stdout = nil
	}
	stderr, err := proc.StderrPipe()
	if err != nil {
		stderr = nil
	}
	return fn(proc, stdout, stderr)
}

const BINARY_CHECK_BUF = 4 * 1024 * 1204

func IsBinaryFile(filepath string) (bool, error) {
	f, err := os.Open(filepath)
	if err != nil {
		return true, err
	}
	defer f.Close()
	buf := make([]byte, BINARY_CHECK_BUF /* 4 MB */)
	n, err := f.Read(buf)
	if err != nil {
		return true, err
	}
	if n < BINARY_CHECK_BUF {
		buf = buf[0:n]
	}
	text := string(buf)
	return strings.Contains(text, "\x00"), nil
}