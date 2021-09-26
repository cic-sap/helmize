package utils

import (
	"bytes"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"io"
	"os"
	"os/exec"
	"regexp"
	"sync"
)

type ShellOptions struct {
	outputStream io.Writer
	copyStdout   bool
	copyStderr   bool
}
type ShellOption func(opts *ShellOptions)

func buildOptions(opts []ShellOption) *ShellOptions {
	opt := &ShellOptions{}
	for _, f := range opts {
		f(opt)
	}
	return opt
}

func WithOutputStream(outputStream io.Writer) ShellOption {
	return func(opts *ShellOptions) {
		opts.outputStream = outputStream
	}
}
func CopyStream(src io.ReadCloser, targets ...io.Writer) func() {

	//filter nil
	ok := make([]io.Writer, 0, len(targets))
	for _, t := range targets {
		//log.Debug().Interface("writer", t).Send()
		if t != nil {
			ok = append(ok, t)
		}
	}
	w := io.MultiWriter(ok...)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		_, err := io.Copy(w, src)
		if err != nil && errors.Is(err, io.EOF) && !errors.Is(err, io.ErrClosedPipe) {
			log.Warn().Err(err).Msg("io.copy get error")
		}
		wg.Done()
	}()

	return func() {
		_ = src.Close()
		wg.Wait()
	}
}

func RunShellOutput(process string, args []string, opts ...ShellOption) (string, string, error) {
	opt := buildOptions(opts)
	buf := bytes.NewBuffer(nil)
	bufErr := bytes.NewBuffer(nil)
	pipeStdoutReader, pipeStdoutWriter := io.Pipe()
	pipeStderrReader, pipeStderrWriter := io.Pipe()
	var stdOutArgs = []io.Writer{
		buf,
		opt.outputStream,
	}
	if opt.copyStdout {
		stdOutArgs = append(stdOutArgs, io.Writer(os.Stdout))
	}

	var stdErrArgs = []io.Writer{
		bufErr,
		opt.outputStream,
	}
	if opt.copyStderr {
		stdErrArgs = append(stdErrArgs, io.Writer(os.Stderr))
	}

	clean1 := CopyStream(pipeStdoutReader, stdOutArgs...)
	clean2 := CopyStream(pipeStderrReader, stdErrArgs...)

	defer func() {
		_ = pipeStderrWriter.Close()
		_ = pipeStdoutWriter.Close()
		clean1()
		clean2()
	}()

	log.Debug().Strs("args", append([]string{process}, FilterPassword(args)...)).Msg("start run shell")

	cmd := exec.Command(process, args...)
	cmd.Stdout = pipeStdoutWriter
	cmd.Stderr = pipeStderrWriter

	err := cmd.Run()
	stdout, stderr := buf.String(), bufErr.String()
	if err != nil {
		log.Error().Err(err).
			Str("stdout", stdout).
			Str("stderr", stderr).
			Msg("run shell " + process + " get error")
	}
	return stdout, stderr, err
}

func FilterPassword(args []string) []string {

	var out = make([]string, 0)
	re := regexp.MustCompile("([\\-]*username=|[\\-]*password=).+")
	for _, v := range args {
		if !re.Match([]byte(v)) {
			out = append(out, v)
			continue
		}
		v2 := string(re.ReplaceAllFunc([]byte(v), func(m []byte) []byte {
			parts := re.FindStringSubmatch(string(m))
			return []byte(parts[1] + "*")
		}))
		out = append(out, v2)
	}
	return out
}
