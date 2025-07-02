// +build ignore

package cmd

import (
	"testing"
)

// This file runs scale tests in isolation
func TestScaleIsolated(t *testing.T) {
	// Run scale command tests
	TestScaleCommand(t)
	TestValidateScaleInputs(t)
	TestBuildScaleRequest(t)
	TestIsValidResourceString(t)
	TestDisplayFunctions(t)
	TestDemoFunctions(t)
}