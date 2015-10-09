package goprofhooks

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime/pprof"
	"runtime/trace"
)

var (
	// if we are writing profiles right now, we keep here the name of the working folder where we write profiles to
	// if we are not writing profiles at the moment, it's just an empty string
	ourProfilesDirectory = ""
)

type writeProfileFxn func(profilesDir string) error

// StartProfiling starts writing profiles and returns path to the directory where they will be placed
// if anything goes wrong corresponding error is returned and no profiling is started
// If writing profiles is in progress, an error will be returned
func StartProfiling() (profilesDirectory string, err error) {
	if traceInProgress() {
		return "", fmt.Errorf("Cannot start profiling, since it's already started")
	}
	profiles, err := ioutil.TempDir("", "profiles")
	if err != nil {
		return "", err
	}
	defer func() {
		// if something is wrong, do cleanup
		if err != nil {
			trace.Stop()
			pprof.StopCPUProfile()
			ourProfilesDirectory = ""
			os.RemoveAll(profiles)
		}
	}()
	if err := startWritingTrace(profiles); err != nil {
		return "", err
	}
	if err := startCPUProfiling(profiles); err != nil {
		return "", err
	}
	ourProfilesDirectory = profiles
	return profiles, nil
}

// StopProfiling stops writing all profiles. Before stopping them it tries to write a heap dump
// to the same folder where other profiles are kept. It returns path to the folder which contains profiling files
// If profiling is not in progress no error will be returned
func StopProfiling() (profilesDirectory string, err error) {
	if !traceInProgress() {
		return "", nil
	}
	defer func() {
		// stop everything when we are finished with writing heap profile
		trace.Stop()
		pprof.StopCPUProfile()
		ourProfilesDirectory = ""
	}()
	return ourProfilesDirectory, writeHeapProfile(ourProfilesDirectory)
}

// ToggleProfiling changes state of writing profiles to the opposite
func ToggleProfiling() (profilesDirectory string, err error) {
	if traceInProgress() {
		return StopProfiling()
	}
	return StartProfiling()
}

func startWritingTrace(profilesDir string) error {
	traceFile, err := os.Create(filepath.Join(profilesDir, "trace"))
	if err != nil {
		return err
	}
	return trace.Start(traceFile)
}

func writeHeapProfile(profilesDir string) error {
	heapProfileFile, err := os.Create(filepath.Join(profilesDir, "heap-profile"))
	if err != nil {
		return err
	}
	defer heapProfileFile.Close()
	return pprof.WriteHeapProfile(heapProfileFile)
}

func startCPUProfiling(profilesDir string) error {
	cpuProfileFile, err := os.Create(filepath.Join(profilesDir, "cpu-profile"))
	if err != nil {
		return err
	}
	return pprof.StartCPUProfile(cpuProfileFile)
}

func traceInProgress() bool {
	return ourProfilesDirectory != ""
}