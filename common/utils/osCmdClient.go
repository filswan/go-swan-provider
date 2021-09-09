package utils

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
)


func ExecOsCmd2Screen(cmdName string, args string) (string ,bool){
	cmd := exec.Command(cmdName, args)

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = io.MultiWriter(os.Stdout, &stdoutBuf)
	cmd.Stderr = io.MultiWriter(os.Stderr, &stderrBuf)

	err := cmd.Run()
	if err != nil {
		fmt.Println(err)
	}

	if len(stderrBuf.Bytes())==0{
		return string(stdoutBuf.Bytes()), true
	}

	return "", false
}

func ExecOsCmd(cmdName string, args string) (string, string){
	cmd := exec.Command(cmdName, args)
	out, err := cmd.CombinedOutput()
	return string(out), err.Error()
}

func ExecOsCmd1(cmdName string, args string) (string, bool){
	cmd := exec.Command(cmdName, args)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		fmt.Println(err)
		return "", false
	}

	if len(stderr.Bytes()) != 0 {
		fmt.Println(string(stderr.Bytes()))
		return "", false
	}

	outStr := string(stdout.Bytes())
	fmt.Printf(outStr)
	return outStr, true
}
