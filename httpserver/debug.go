package httpserver

import (
	"bytes"
	"fmt"
	"io"
	"os"
)

func doSafe(f func()) (err error) {
	defer func() {
		if v := recover(); v != nil {
			err = fmt.Errorf("panic: %v", v)
		}
	}()

	f()

	return nil
}

func doDebug(f func()) (stdout string, stderr string, err error) {
	defer func() {
		if v := recover(); v != nil {
			err = fmt.Errorf("panic: %v", v)
		}
	}()

	// keep backup of the real file
	originStdout := os.Stdout
	originStderr := os.Stderr

	// Restore original file
	defer func() {
		os.Stdout = originStdout
		os.Stderr = originStderr
	}()

	// Create pipe to create reader & writer
	stdoutPipeReader, stdoutPipeWriter, err := os.Pipe()
	if err != nil {
		panic(err)
	}
	defer stdoutPipeWriter.Close()
	stderrPipeReader, stderrPipeWriter, err := os.Pipe()
	if err != nil {
		panic(err)
	}
	defer stderrPipeWriter.Close()

	// Connect file to writer side of pipe
	os.Stdout = stdoutPipeWriter
	os.Stderr = stderrPipeWriter

	// Create MultiWriter to write to buffer and file at the same time
	var (
		stdoutBuf bytes.Buffer
		stderrBuf bytes.Buffer
	)
	stdoutMultiWriter := io.MultiWriter(&stdoutBuf, originStdout)
	stderrMultiWriter := io.MultiWriter(&stderrBuf, originStderr)

	// copy the output in a separate goroutine so printing can't block indefinitely
	stdoutErrCh := make(chan error, 1)
	go func() {
		if _, err := io.Copy(stdoutMultiWriter, stdoutPipeReader); err != nil {
			stdoutErrCh <- err
		}
		stdoutErrCh <- nil
	}()
	go func() {
		if _, err := io.Copy(stderrMultiWriter, stderrPipeReader); err != nil {
			stdoutErrCh <- err
		}
		stdoutErrCh <- nil
	}()

	f()

	if err := stdoutPipeWriter.Close(); err != nil {
		panic(err)
	}

	if err := stderrPipeWriter.Close(); err != nil {
		panic(err)
	}

	if err := <-stdoutErrCh; err != nil {
		panic(err)
	}

	if err := <-stdoutErrCh; err != nil {
		panic(err)
	}

	return stdoutBuf.String(), stderrBuf.String(), nil
}
