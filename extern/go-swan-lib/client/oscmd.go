package client

import (
	"bytes"
	"errors"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/filswan/go-swan-lib/logs"
)

const SHELL_TO_USE = "bash"

func ExecOsCmd2Screen(cmdStr string, checkStdErr bool) (string, error) {
	out, err := ExecOsCmdBase(cmdStr, true, checkStdErr)
	return out, err
}

func ExecOsCmd(cmdStr string, checkStdErr bool) (string, error) {
	out, err := ExecOsCmdBase(cmdStr, false, checkStdErr)
	return out, err
}

func ExecOsCmdBase(cmdStr string, out2Screen bool, checkStdErr bool) (string, error) {
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
	if err != nil || len(stderrBuf.String()) != 0 {
		errs := []string{}
		errs = append(errs, stderrBuf.String())
		errs = append(errs, err.Error())
		errMsg := strings.Join(errs, ",")

		outErr := errors.New(errMsg)
		logs.GetLogger().Error(outErr)
		return "", outErr
	}

	return stdoutBuf.String(), nil
}
