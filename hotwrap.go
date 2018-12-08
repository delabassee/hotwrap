package main

import (
	"context"
	fdk "github.com/fnproject/fdk-go"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
)

func main() {

	if len(os.Args) < 2 {
		log.Fatal("Failed to start hotwrap, no command specified in arguments ")
	}

	if os.Getenv("HOTWRAP_VERBOSE") != "" {
	}

	cmd := os.Args[1]
	var args []string
	if len(os.Args) > 1 {
		args = os.Args[2:]
	}

	fdk.Handle(&hotWrap{
		cmd:  cmd,
		args: args,
		env:  os.Environ(),
	})

}

type hotWrap struct {
	verbose bool
	cmd     string
	args    []string
	env     []string
}

func (hw *hotWrap) logf(fmt string, args ...interface{}) {
	if hw.verbose {
		log.Printf(fmt, args...)
	}
}

func (hw *hotWrap) Serve(ctx context.Context, r io.Reader, w io.Writer) {

	rctx := fdk.GetContext(ctx)
	hdr := rctx.Header()

	// Each env entry is of the form "key=value".
	baseEnv := hw.env

	// Map custom Fn-Http-H-* headers to FN_HEADER_* env variables
	// ToDo: check that map custom Fn-Http-H-* headers to FN_HEADER_* env variables is still applicable
	// ToDo: check Lowercasing behaviour for K
	//   curl -H "lower:lower" -> FN_HEADER_Lower='lower'
	//   curl -H "UPER:UPER" -> FN_HEADER_Uper='UPER'
	// See https://github.com/fnproject/fdk-java/blob/eac3fffe1decde276536821c2924b425a4589e7d/api/src/main/java/com/fnproject/fn/api/Headers.java#L15-L23
	for k, vs := range hdr {
		switch {
		case strings.HasPrefix(k, "Fn-Http-H-Accept"):
		case strings.HasPrefix(k, "Fn-Http-H-User-Agent"):
		case strings.HasPrefix(k, "Fn-Http-H-Content-Length"):
		case strings.HasPrefix(k, "Fn-Http-H-"):
			envVar := "FN_HEADER_" + strings.TrimPrefix(k, "Fn-Http-H-") + "=" + vs[0]
			baseEnv = append(baseEnv, envVar)
		default:
		}
	}

	cmd := exec.Command(hw.cmd, hw.args...)

	cmd.Env = baseEnv
	cmd.Stdout = w
	cmd.Stdin = r

	stderr, err := cmd.StderrPipe()

	if err != nil {
		log.Fatalf("Failed to open stderr pipe %s", err)
	}

	go func() {
		io.Copy(os.Stderr, stderr)
	}()

	err = cmd.Start()
	if err != nil {
		log.Fatalf("Failed to start command %s", err)
	}

	err = cmd.Wait()

	if ee, ok := err.(*exec.ExitError); ok {
		log.Printf("Command %s exited with status %s", hw.cmd, ee.ProcessState)
		fdk.WriteStatus(w, 500)
	}

}
