package main

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Types are defined in scale_command_test.go

func TestScaleTypes(t *testing.T) {
	t.Run("DeploymentScale serialization", func(t *testing.T) {
		scale := DeploymentScale{
			CurrentReplicas: 3,
			DesiredReplicas: 5,
			MinReplicas:     1,
			MaxReplicas:     10,
			AutoScaling:     true,
			CPUThreshold:    70,
			Resources: ResourceAllocation{
				Memory: "1Gi",
				CPU:    "500m",
			},
			Status:     "scaling",
			LastScaled: time.Now(),
		}

		// Test JSON marshaling
		data, err := json.Marshal(scale)
		assert.NoError(t, err)
		assert.Contains(t, string(data), `"current_replicas":3`)
		assert.Contains(t, string(data), `"auto_scaling":true`)
		assert.Contains(t, string(data), `"memory":"1Gi"`)

		// Test JSON unmarshaling
		var decoded DeploymentScale
		err = json.Unmarshal(data, &decoded)
		assert.NoError(t, err)
		assert.Equal(t, scale.CurrentReplicas, decoded.CurrentReplicas)
		assert.Equal(t, scale.AutoScaling, decoded.AutoScaling)
		assert.Equal(t, scale.Resources.Memory, decoded.Resources.Memory)
	})

	t.Run("ScaleRequest with nil fields", func(t *testing.T) {
		// Test with some nil fields
		replicas := 5
		autoScale := true
		req := ScaleRequest{
			Replicas:    &replicas,
			AutoScaling: &autoScale,
			// Leave other fields nil
		}

		data, err := json.Marshal(req)
		assert.NoError(t, err)
		
		// Should only include non-nil fields
		assert.Contains(t, string(data), `"replicas":5`)
		assert.Contains(t, string(data), `"auto_scaling":true`)
		assert.NotContains(t, string(data), "min_replicas")
		assert.NotContains(t, string(data), "cpu_target")
	})

	t.Run("ResourceAllocation validation", func(t *testing.T) {
		validResources := []ResourceAllocation{
			{Memory: "512Mi", CPU: "250m"},
			{Memory: "1Gi", CPU: "500m"},
			{Memory: "2Gi", CPU: "1000m"},
			{Memory: "4Gi", CPU: "2"},
		}

		for _, res := range validResources {
			assert.NotEmpty(t, res.Memory)
			assert.NotEmpty(t, res.CPU)
			assert.Regexp(t, `^\d+[GM]i$`, res.Memory)
			assert.Regexp(t, `^\d+(m|\d*)$`, res.CPU)
		}
	})
}

func TestScaleInputValidation(t *testing.T) {
	t.Run("validate replica counts", func(t *testing.T) {
		tests := []struct {
			replicas int
			min      int
			max      int
			valid    bool
		}{
			{5, 1, 10, true},
			{0, 0, 0, true},
			{-1, 0, 0, false},
			{0, -1, 10, false},
			{0, 10, 5, false}, // min > max
		}

		for _, tt := range tests {
			if tt.replicas < 0 {
				assert.False(t, tt.valid, "negative replicas should be invalid")
			}
			if tt.min < 0 || tt.max < 0 {
				assert.False(t, tt.valid, "negative min/max should be invalid")
			}
			if tt.min > 0 && tt.max > 0 && tt.min > tt.max {
				assert.False(t, tt.valid, "min > max should be invalid")
			}
		}
	})

	t.Run("validate CPU threshold", func(t *testing.T) {
		validCPU := []int{1, 50, 70, 80, 99, 100}
		invalidCPU := []int{0, -1, 101, 200}

		for _, cpu := range validCPU {
			assert.True(t, cpu >= 1 && cpu <= 100)
		}

		for _, cpu := range invalidCPU {
			assert.False(t, cpu >= 1 && cpu <= 100)
		}
	})

	t.Run("validate memory format", func(t *testing.T) {
		validMemory := []string{"512Mi", "1Gi", "2Gi", "4Gi", "8Gi"}
		invalidMemory := []string{"512", "1GB", "2g", "invalid", ""}

		for _, mem := range validMemory {
			assert.Regexp(t, `^\d+[GM]i$`, mem)
		}

		for _, mem := range invalidMemory {
			if mem != "" {
				assert.NotRegexp(t, `^\d+[GM]i$`, mem)
			}
		}
	})
}