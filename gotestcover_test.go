package main

import (
	"github.com/stretchr/testify/assert"
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
		"-timeout=15",
		"-covermode=atomic",
		"-parallelpackages=2",
		"-coverprofile=cover.out",
	}

	err := parseFlags()

	assert.Nil(t, err)
	assert.True(t, flagVerbose)
	assert.True(t, flagA)
	assert.True(t, flagX)
	assert.True(t, flagRace)
	assert.Equal(t, "4", flagCPU)
	assert.Equal(t, "2", flagParallel)
	assert.Equal(t, "abc", flagRun)
	assert.True(t, flagShort)
	assert.Equal(t, "15", flagTimeout)
	assert.Equal(t, "atomic", flagCoverMode)
	assert.Equal(t, 2, flagParallelPackages)
	assert.Equal(t, "cover.out", flagCoverProfile)
}
