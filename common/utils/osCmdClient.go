package utils

import (
	"bytes"
	"io"
	"os"
	"os/exec"
)

const SHELL_TO_USE = "bash"

func ExecOsCmd2Screen(cmdName string) (string ,string){
	var stdoutBuf bytes.Buffer
	var stderrBuf bytes.Buffer

	cmd := exec.Command(SHELL_TO_USE, "-c", cmdName)

	cmd.Stdout = io.MultiWriter(os.Stdout, &stdoutBuf)
	cmd.Stderr = io.MultiWriter(os.Stderr, &stderrBuf)

	err := cmd.Run()
	if err != nil {
		//fmt.Println(err)
		return "", err.Error()
	}

	if len(stderrBuf.Bytes()) != 0 {
		//fmt.Println(string(stderr.Bytes()))
		return "", string(stderrBuf.Bytes())
	}

	return string(stdoutBuf.Bytes()), ""
}

func ExecOsCmd(cmdName string) (string, string) {
	var stdoutBuf bytes.Buffer
	var stderrBuf bytes.Buffer

	cmd := exec.Command(SHELL_TO_USE, "-c", cmdName)

	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf
	err := cmd.Run()
	if err != nil {
		//fmt.Println(err)
		return "", err.Error()
	}

	if len(stderrBuf.Bytes()) != 0 {
		//fmt.Println(string(stderr.Bytes()))
		return "", string(stderrBuf.Bytes())
	}

	outStr := string(stdoutBuf.Bytes())
	//fmt.Printf(outStr)
	return outStr, ""
}
