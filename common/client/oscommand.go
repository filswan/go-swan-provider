package client

import (
	"bytes"
	"errors"
	"io"
	"os"
	"os/exec"
	"swan-provider/logs"
)

const SHELL_TO_USE = "bash"

func ExecOsCmd2Screen(cmdStr string) (string, error) {
	out, err := ExecOsCmdBase(cmdStr, true)
	return out, err
}

func ExecOsCmd(cmdStr string) (string, error) {
	out, err := ExecOsCmdBase(cmdStr, false)
	return out, err
}

func ExecOsCmdBase(cmdStr string, out2Screen bool) (string, error) {
	var stdoutBuf bytes.Buffer
	var stderrBuf bytes.Buffer

	cmd := exec.Command(SHELL_TO_USE, "-c", cmdStr)

	if out2Screen {
		cmd.Stdout = io.MultiWriter(os.Stdout, &stdoutBuf)
		cmd.Stderr = io.MultiWriter(os.Stderr, &stderrBuf)
	} else {
		cmd.Stdout = &stdoutBuf
		cmd.Stderr = &stderrBuf
	}

	err := cmd.Run()
	if err != nil {
		logs.GetLogger().Error(cmdStr, err)
		return "", err
	}

	if len(stderrBuf.Bytes()) != 0 {
		outErr := errors.New(stderrBuf.String())
		logs.GetLogger().Error(cmdStr, outErr)
		return "", outErr
	}

	return stdoutBuf.String(), nil
}
