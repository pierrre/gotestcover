package main

import (
	"os"
	"testing"
)

func TestParseFlags(t *testing.T) {
	os.Args = []string{"gotestcover",
		"-v",
		"-a",
		"-x",
		"-race",
		"-cpu=4",
		"-parallel=2",
		"-run=abc",
		"-short",
		"-timeout=15s",
		"-covermode=atomic",
		"-parallelpackages=2",
		"-coverprofile=cover.out",
	}

	err := parseFlags()

	if err != nil {
		t.Fatal(err)
	}

	if !flagVerbose {
		t.Errorf("flagVerbose should be set to true")
	}

	if !flagA {
		t.Errorf("flagA should be set to true")
	}

	if !flagX {
		t.Errorf("flagX should be set to true")
	}

	if !flagRace {
		t.Errorf("flagRace should be set to true")
	}

	if flagCPU != "4" {
		t.Errorf("flagCPU is not equal to 4, got %s", flagCPU)
	}

	if flagParallel != "2" {
		t.Errorf("flagCPU is not equal to 2, got %s", flagParallel)
	}

	if flagRun != "abc" {
		t.Errorf("flagRun is not equal to 'abc', got %s", flagRun)
	}

	if !flagShort {
		t.Errorf("flagShort should be set to true")
	}

	if flagTimeout != "15s" {
		t.Errorf("flagTimeout is not equal to '15s', got %s", flagTimeout)
	}

	if flagCoverMode != "atomic" {
		t.Errorf("flagCoverMode is not equal to 'atomic', got %s", flagCoverMode)
	}

	if flagParallelPackages != 2 {
		t.Errorf("flagParallelPackages is not equal to '2', got %d", flagParallelPackages)
	}

	if flagCoverProfile != "cover.out" {
		t.Errorf("flagCoverProfile is not equal to 'cover.out', got %s", flagCoverProfile)
	}
}
