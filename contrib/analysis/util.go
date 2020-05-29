package analysis

import (
	"os/exec"
	"sync"
	"os"
	"io"
	"bufio"
	"strings"
)

func doExec(cmd string, fn func (line string)) error {
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
	stderr, err := proc.StderrPipe()
	if err = proc.Start(); err != nil {
		return err
	}
	listener := &sync.WaitGroup{}
	listener.Add(2)
	go watchOutput(proc, listener, stdout, fn)
	go watchOutput(proc, listener, stderr, fn)
	listener.Wait()
	proc.Wait()
	return nil
}

func watchOutput(proc *exec.Cmd, listener *sync.WaitGroup, stream io.ReadCloser, fn func (line string)) {
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
