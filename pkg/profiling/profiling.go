// Package profiling provides utilities for performance profiling
package profiling

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"time"

	"github.com/cpuix/multigit/pkg/logger"
)

// ProfileType represents the type of profile to collect
type ProfileType string

const (
	// CPUProfile represents CPU profiling
	CPUProfile ProfileType = "cpu"
	// MemoryProfile represents memory profiling
	MemoryProfile ProfileType = "memory"
	// BlockProfile represents block profiling
	BlockProfile ProfileType = "block"
	// MutexProfile represents mutex profiling
	MutexProfile ProfileType = "mutex"
)

// StartCPUProfile starts CPU profiling and returns a function to stop it
func StartCPUProfile(outputDir string) (stopFunc func(), err error) {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	cpuProfilePath := filepath.Join(outputDir, fmt.Sprintf("cpu_%s.prof", time.Now().Format("20060102_150405")))
	f, err := os.Create(cpuProfilePath)
	if err != nil {
		return nil, fmt.Errorf("could not create CPU profile: %w", err)
	}

	logger.Infof("CPU profiling enabled, writing to: %s", cpuProfilePath)
	
	if err := pprof.StartCPUProfile(f); err != nil {
		f.Close()
		return nil, fmt.Errorf("could not start CPU profile: %w", err)
	}

	stopFunc = func() {
		pprof.StopCPUProfile()
		f.Close()
		logger.Infof("CPU profile written to: %s", cpuProfilePath)
	}

	return stopFunc, nil
}

// WriteHeapProfile writes a heap profile to the specified directory
func WriteHeapProfile(outputDir string) error {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	heapProfilePath := filepath.Join(outputDir, fmt.Sprintf("heap_%s.prof", time.Now().Format("20060102_150405")))
	f, err := os.Create(heapProfilePath)
	if err != nil {
		return fmt.Errorf("could not create memory profile: %w", err)
	}
	defer f.Close()

	runtime.GC() // Get up-to-date statistics
	if err := pprof.WriteHeapProfile(f); err != nil {
		return fmt.Errorf("could not write memory profile: %w", err)
	}

	logger.Infof("Memory profile written to: %s", heapProfilePath)
	return nil
}

// WriteBlockProfile writes a block profile to the specified directory
func WriteBlockProfile(outputDir string) error {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	blockProfilePath := filepath.Join(outputDir, fmt.Sprintf("block_%s.prof", time.Now().Format("20060102_150405")))
	f, err := os.Create(blockProfilePath)
	if err != nil {
		return fmt.Errorf("could not create block profile: %w", err)
	}
	defer f.Close()

	runtime.SetBlockProfileRate(1)
	defer runtime.SetBlockProfileRate(0)

	if err := pprof.Lookup("block").WriteTo(f, 0); err != nil {
		return fmt.Errorf("could not write block profile: %w", err)
	}

	logger.Infof("Block profile written to: %s", blockProfilePath)
	return nil
}

// WriteMutexProfile writes a mutex profile to the specified directory
func WriteMutexProfile(outputDir string) error {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	mutexProfilePath := filepath.Join(outputDir, fmt.Sprintf("mutex_%s.prof", time.Now().Format("20060102_150405")))
	f, err := os.Create(mutexProfilePath)
	if err != nil {
		return fmt.Errorf("could not create mutex profile: %w", err)
	}
	defer f.Close()

	runtime.SetMutexProfileFraction(1)
	defer runtime.SetMutexProfileFraction(0)

	if err := pprof.Lookup("mutex").WriteTo(f, 0); err != nil {
		return fmt.Errorf("could not write mutex profile: %w", err)
	}

	logger.Infof("Mutex profile written to: %s", mutexProfilePath)
	return nil
}

// CollectProfiles collects all available profiles and writes them to the specified directory
func CollectProfiles(outputDir string) (cleanupFunc func(), err error) {
	var cleanups []func()
	cleanupFunc = func() {
		for _, f := range cleanups {
			f()
		}
	}

	// Handle any errors during cleanup
	defer func() {
		if err != nil {
			cleanupFunc()
		}
	}()

	// Start CPU profiling
	stopCPUProfile, err := StartCPUProfile(outputDir)
	if err != nil {
		return nil, fmt.Errorf("failed to start CPU profile: %w", err)
	}
	cleanups = append(cleanups, stopCPUProfile)

	// Schedule memory profile collection on exit
	cleanups = append(cleanups, func() {
		if err := WriteHeapProfile(outputDir); err != nil {
			logger.Errorf("Failed to write heap profile: %v", err)
		}
	})

	// Schedule block profile collection on exit
	cleanups = append(cleanups, func() {
		if err := WriteBlockProfile(outputDir); err != nil {
			logger.Errorf("Failed to write block profile: %v", err)
		}
	})

	// Schedule mutex profile collection on exit
	cleanups = append(cleanups, func() {
		if err := WriteMutexProfile(outputDir); err != nil {
			logger.Errorf("Failed to write mutex profile: %v", err)
		}
	})

	return cleanupFunc, nil
}
