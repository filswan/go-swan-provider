package utils

import (
	"bytes"
	"io"
	"os"
	"os/exec"
)

const SHELL_TO_USE = "bash"


func ExecOsCmd2Screen(cmdName string) (string ,string){
	out, err := ExecOsCmdBase(cmdName, true)
	return out, err
}

func ExecOsCmd(cmdName string) (string, string) {
	out, err := ExecOsCmdBase(cmdName, false)
	return out, err
}

func ExecOsCmdBase(cmdName string, out2Screen bool) (string ,string){
	var stdoutBuf bytes.Buffer
	var stderrBuf bytes.Buffer

	cmd := exec.Command(SHELL_TO_USE, "-c", cmdName)

	if out2Screen {
		cmd.Stdout = io.MultiWriter(os.Stdout, &stdoutBuf)
		cmd.Stderr = io.MultiWriter(os.Stderr, &stderrBuf)
	} else {
		cmd.Stdout = &stdoutBuf
		cmd.Stderr = &stderrBuf
	}

	err := cmd.Run()
	if err != nil {
		outErr := err.Error()
		logger.Error(cmdName, outErr)
		return "", outErr
	}

	if len(stderrBuf.Bytes()) != 0 {
		outErr := string(stderrBuf.Bytes())
		logger.Error(cmdName, outErr)
		return "", outErr
	}

	return string(stdoutBuf.Bytes()), ""
}
