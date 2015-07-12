package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"syscall"
)

var (
	flagVerbose   bool
	flagA         bool
	flagX         bool
	flagRace      bool
	flagCPU       string
	flagParallel  string
	flagRun       string
	flagShort     bool
	flagTimeout   string
	flagCoverMode = "set"
	flagHTML      string
)

func main() {
	parseFlags()
	err := run()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func parseFlags() {
	flag.BoolVar(&flagVerbose, "v", flagVerbose, "see `go test` help")
	flag.BoolVar(&flagA, "a", flagA, "see `go build` help")
	flag.BoolVar(&flagX, "x", flagX, "see `go build` help")
	flag.BoolVar(&flagRace, "race", flagRace, "see `go build` help")
	flag.StringVar(&flagCPU, "cpu", flagCPU, "see `go test` help")
	flag.StringVar(&flagParallel, "parallel", flagParallel, "see `go test` help")
	flag.StringVar(&flagRun, "run", flagRun, "see `go test` help")
	flag.BoolVar(&flagShort, "short", flagShort, "see `go test` help")
	flag.StringVar(&flagTimeout, "timeout", flagTimeout, "see `go test` help")
	flag.StringVar(&flagCoverMode, "covermode", flagCoverMode, "see `go test` help")
	flag.StringVar(&flagHTML, "html", flagHTML, "generate HTML representation of coverage profile")
	flag.Parse()
}

func run() error {
	ps, err := getPackages()
	if err != nil {
		return err
	}
	cov, err := runAllPackageTests(ps, func(out string) {
		fmt.Print(out)
	})
	if err != nil {
		return err
	}
	covRes, err := generateCoverResult(cov)
	if err != nil {
		return err
	}
	fmt.Printf("%s", covRes)
	return nil
}

func getPackages() ([]string, error) {
	cmdArgs := []string{"list"}
	cmdArgs = append(cmdArgs, flag.Args()...)
	cmdOut, err := runGoCommand(cmdArgs...)
	if err != nil {
		return nil, err
	}
	var ps []string
	sc := bufio.NewScanner(bytes.NewReader(cmdOut))
	for sc.Scan() {
		ps = append(ps, sc.Text())
	}
	return ps, nil
}

func runAllPackageTests(ps []string, pf func(string)) ([]byte, error) {
	covBuf := new(bytes.Buffer)
	fmt.Fprintf(covBuf, "mode: %s\n", flagCoverMode)
	for _, p := range ps {
		out, cov, err := runPackageTests(p)
		if err != nil {
			return nil, err
		}
		pf(out)
		covBuf.Write(cov)
	}
	return covBuf.Bytes(), nil
}

func runPackageTests(p string) (out string, cov []byte, err error) {
	coverFile, err := tempFile()
	if err != nil {
		return "", nil, err
	}
	defer coverFile.Close()
	defer os.Remove(coverFile.Name())
	var args []string
	args = append(args, "test")
	if flagVerbose {
		args = append(args, "-v")
	}
	if flagA {
		args = append(args, "-a")
	}
	if flagX {
		args = append(args, "-x")
	}
	if flagRace {
		args = append(args, "-race")
	}
	if flagCPU != "" {
		args = append(args, "-cpu", flagCPU)
	}
	if flagParallel != "" {
		args = append(args, "-parallel", flagParallel)
	}
	if flagRun != "" {
		args = append(args, "-run", flagRun)
	}
	if flagShort {
		args = append(args, "-short")
	}
	if flagTimeout != "" {
		args = append(args, "-timeout", flagTimeout)
	}
	args = append(args, "-cover")
	if flagCoverMode != "" {
		args = append(args, "-covermode", flagCoverMode)
	}
	args = append(args, "-coverprofile", coverFile.Name())
	args = append(args, p)
	cmdOut, err := runGoCommandExpectExitError(1, args...)
	if err != nil {
		return "", nil, err
	}
	cov, err = ioutil.ReadAll(coverFile)
	if err != nil {
		return "", nil, err
	}
	cov = removeFirstLine(cov)
	return string(cmdOut), cov, nil
}

func generateCoverResult(cov []byte) ([]byte, error) {
	coverFile, err := tempFile()
	if err != nil {
		return nil, err
	}
	defer coverFile.Close()
	defer os.Remove(coverFile.Name())
	err = ioutil.WriteFile(coverFile.Name(), cov, 0)
	if err != nil {
		return nil, err
	}
	covRes, err := runGoCommand("tool", "cover", "-func", coverFile.Name())
	if err != nil {
		return nil, err
	}
	if flagHTML != "" {
		_, err := runGoCommand("tool", "cover", "-html", coverFile.Name(), "-o", flagHTML)
		if err != nil {
			return nil, err
		}
	}
	return covRes, nil
}

func runGoCommand(args ...string) ([]byte, error) {
	return runGoCommandExpectExitError(0, args...)
}

func runGoCommandExpectExitError(expectedStatus int, args ...string) ([]byte, error) {
	cmd := exec.Command("go", args...)
	out, err := cmd.CombinedOutput()
	if err == nil {
		return out, nil
	}
	if err, ok := err.(*exec.ExitError); ok {
		if status, ok := err.Sys().(syscall.WaitStatus); ok {
			if status.ExitStatus() == expectedStatus {
				return out, nil
			}
		}
	}
	return nil, fmt.Errorf("command %s: %s\n%s", cmd.Args, err, out)
}

func removeFirstLine(b []byte) []byte {
	out := new(bytes.Buffer)
	sc := bufio.NewScanner(bytes.NewReader(b))
	firstLine := true
	for sc.Scan() {
		if firstLine {
			firstLine = false
			continue
		}
		fmt.Fprintf(out, "%s\n", sc.Bytes())
	}
	return out.Bytes()
}

func tempFile() (*os.File, error) {
	return ioutil.TempFile("", "gotestcover-")
}
