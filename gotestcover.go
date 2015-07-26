// Package gotestcover provides multiple packages support for Go test cover.
package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"runtime"
	"sync"
)

var (
	flagVerbose          bool
	flagA                bool
	flagX                bool
	flagRace             bool
	flagCPU              string
	flagParallel         string
	flagRun              string
	flagShort            bool
	flagTimeout          string
	flagCoverMode        string
	flagCoverProfile     string
	flagParallelPackages = runtime.GOMAXPROCS(0)
)

func main() {
	err := run()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func run() error {
	err := parseFlags()
	if err != nil {
		return err
	}
	ps, err := getPackages()
	if err != nil {
		return err
	}
	cov := runAllPackageTests(ps, func(out string) {
		fmt.Print(out)
	})
	err = writeCoverProfile(cov)
	if err != nil {
		return err
	}
	return nil
}

func parseFlags() error {
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
	flag.StringVar(&flagCoverProfile, "coverprofile", flagCoverProfile, "see `go test` help")
	flag.IntVar(&flagParallelPackages, "parallelpackages", flagParallelPackages, "Number of package test run in parallel")
	flag.Parse()
	if flagCoverProfile == "" {
		return fmt.Errorf("flag coverprofile must be set")
	}
	if flagParallelPackages < 1 {
		return fmt.Errorf("flag parallelpackages must be greater than or equal to 1")
	}
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

func runAllPackageTests(ps []string, pf func(string)) []byte {
	pch := make(chan string)
	type res struct {
		out string
		cov []byte
		err error
	}
	resch := make(chan res)
	wg := new(sync.WaitGroup)
	wg.Add(flagParallelPackages)
	go func() {
		for _, p := range ps {
			pch <- p
		}
		close(pch)
		wg.Wait()
		close(resch)
	}()
	for i := 0; i < flagParallelPackages; i++ {
		go func() {
			for p := range pch {
				out, cov, err := runPackageTests(p)
				resch <- res{
					out: out,
					cov: cov,
					err: err,
				}
			}
			wg.Done()
		}()
	}
	var cov []byte
	for r := range resch {
		if r.err == nil {
			pf(r.out)
			cov = append(cov, r.cov...)
		} else {
			pf(r.err.Error())
		}
	}
	return cov
}

func runPackageTests(p string) (out string, cov []byte, err error) {
	coverFile, err := ioutil.TempFile("", "gotestcover-")
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
	cmdOut, err := runGoCommand(args...)
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

func writeCoverProfile(cov []byte) error {
	covBuf := new(bytes.Buffer)
	coverMode := flagCoverMode
	if coverMode == "" {
		if flagRace {
			coverMode = "atomic"
		} else {
			coverMode = "set"
		}
	}
	fmt.Fprintf(covBuf, "mode: %s\n", coverMode)
	covBuf.Write(cov)
	return ioutil.WriteFile(flagCoverProfile, covBuf.Bytes(), os.FileMode(0644))
}

func runGoCommand(args ...string) ([]byte, error) {
	cmd := exec.Command("go", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("command %s: %s\n%s", cmd.Args, err, out)
	}
	return out, nil
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
