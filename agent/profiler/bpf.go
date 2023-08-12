package profiler

import (
	"fmt"
	"github.com/simplexiengage/kubectl-flame/agent/details"
	"github.com/simplexiengage/kubectl-flame/agent/utils"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	kernelSourcesDir         = "/usr/src/kernel-source/"
	profilerLocation         = "/app/bcc-profiler/profile"
	rawProfilerOutputFile    = "/tmp/raw_profile.txt"
	flameGraphScriptLocation = "/app/FlameGraph/flamegraph.pl"
	flameGraphOutputLocation = "/tmp/flamegraph.svg"
)

type BpfProfiler struct{}

func (b *BpfProfiler) SetUp(job *details.ProfilingJob) error {
	exitCode, kernelVersion, err := utils.ExecuteCommand(exec.Command("uname", "-r"))
	if err != nil {
		return fmt.Errorf("failed to get kernel version, exit code: %d, error: %s", exitCode, err)
	}

	expectedSourcesLocation, err := os.Readlink(fmt.Sprintf("/lib/modules/%s/build",
		strings.TrimSuffix(kernelVersion, "\n")))
	if err != nil {
		return fmt.Errorf("failed to read source link, error: %s", err)
	}

	return b.moveSources(expectedSourcesLocation)
}

func (b *BpfProfiler) Invoke(job *details.ProfilingJob) error {
	err := b.runProfiler(job)
	if err != nil {
		return fmt.Errorf("profiling failed: %s", err)
	}

	err = b.generateFlameGraph()
	if err != nil {
		return fmt.Errorf("flamegraph generation failed: %s", err)
	}

	return utils.PublishFlameGraph(flameGraphOutputLocation)
}

func (b *BpfProfiler) runProfiler(job *details.ProfilingJob) error {
	pid, err := utils.FindProcessId(job)
	if err != nil {
		return err
	}

	f, err := os.Create(rawProfilerOutputFile)
	if err != nil {
		return err
	}
	defer f.Close()

	duration := strconv.Itoa(int(job.Duration.Seconds()))
	profileCmd := exec.Command(profilerLocation, "-df", "-p", pid, duration)
	profileCmd.Stdout = f

	return profileCmd.Run()
}

func (b *BpfProfiler) generateFlameGraph() error {
	inputFile, err := os.Open(rawProfilerOutputFile)
	if err != nil {
		return err
	}
	defer inputFile.Close()

	outputFile, err := os.Create(flameGraphOutputLocation)
	if err != nil {
		return err
	}
	defer outputFile.Close()

	flameGraphCmd := exec.Command(flameGraphScriptLocation)
	flameGraphCmd.Stdin = inputFile
	flameGraphCmd.Stdout = outputFile

	return flameGraphCmd.Run()
}

func (b *BpfProfiler) moveSources(target string) error {
	parent, _ := filepath.Split(target)
	err := os.MkdirAll(parent, os.ModePerm)
	if err != nil {
		return err
	}

	_, _, err = utils.ExecuteCommand(exec.Command("mv", kernelSourcesDir, target))
	if err != nil {
		return fmt.Errorf("failed moving source files, error: %s, tried to move to: %s", err, target)
	}

	return nil
}
