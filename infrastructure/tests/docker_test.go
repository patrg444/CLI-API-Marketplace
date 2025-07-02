package test

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/gruntwork-io/terratest/modules/docker"
	"github.com/gruntwork-io/terratest/modules/shell"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDockerImages(t *testing.T) {
	t.Parallel()

	// Test backend API Docker image
	t.Run("validate backend API image", func(t *testing.T) {
		tag := fmt.Sprintf("api-marketplace-backend:test-%d", time.Now().Unix())
		buildOptions := &docker.BuildOptions{
			Tags: []string{tag},
			BuildArgs: []string{
				"NODE_ENV=test",
				"BUILD_VERSION=test",
			},
		}

		// Build the image
		docker.Build(t, "../docker/backend", buildOptions)
		defer docker.DeleteImage(t, tag)

		// Validate image size (should be optimized)
		cli, err := client.NewClientWithOpts(client.FromEnv)
		require.NoError(t, err)

		imageInfo, _, err := cli.ImageInspectWithRaw(context.Background(), tag)
		require.NoError(t, err)

		// Image should be less than 500MB
		assert.Less(t, imageInfo.Size, int64(500*1024*1024))

		// Validate multi-stage build worked (no build dependencies)
		assert.NotContains(t, imageInfo.Config.Env, "npm_config_cache")

		// Test running the container
		opts := &docker.RunOptions{
			Remove: true,
			Detach: true,
			Name:   "api-backend-test",
			EnvironmentVariables: []string{
				"PORT=3000",
				"NODE_ENV=test",
			},
			Ports: map[string]string{
				"3000/tcp": "3000",
			},
		}

		containerId := docker.Run(t, tag, opts)
		defer docker.Stop(t, []string{containerId}, &docker.StopOptions{})

		// Wait for container to be ready
		time.Sleep(5 * time.Second)

		// Test health endpoint
		resp, err := http.Get("http://localhost:3000/health")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	// Test frontend Docker image
	t.Run("validate frontend image", func(t *testing.T) {
		tag := fmt.Sprintf("api-marketplace-frontend:test-%d", time.Now().Unix())
		buildOptions := &docker.BuildOptions{
			Tags: []string{tag},
			BuildArgs: []string{
				"NEXT_PUBLIC_API_URL=http://localhost:3000",
			},
		}

		docker.Build(t, "../docker/frontend", buildOptions)
		defer docker.DeleteImage(t, tag)

		// Validate nginx configuration is included
		output := docker.Run(t, tag, &docker.RunOptions{
			Command: []string{"ls", "/etc/nginx/conf.d/"},
			Remove:  true,
		})
		assert.Contains(t, output, "default.conf")

		// Validate static assets are built
		output = docker.Run(t, tag, &docker.RunOptions{
			Command: []string{"ls", "/usr/share/nginx/html/_next/static"},
			Remove:  true,
		})
		assert.Contains(t, output, "chunks")
		assert.Contains(t, output, "css")
	})

	// Test database initialization image
	t.Run("validate database init image", func(t *testing.T) {
		tag := fmt.Sprintf("api-marketplace-db-init:test-%d", time.Now().Unix())
		buildOptions := &docker.BuildOptions{
			Tags: []string{tag},
		}

		docker.Build(t, "../docker/db-init", buildOptions)
		defer docker.DeleteImage(t, tag)

		// Validate migration scripts are included
		output := docker.Run(t, tag, &docker.RunOptions{
			Command: []string{"ls", "/docker-entrypoint-initdb.d/"},
			Remove:  true,
		})
		assert.Contains(t, output, "01_schema.sql")
		assert.Contains(t, output, "02_functions.sql")
		assert.Contains(t, output, "03_seed_data.sql")
	})
}

func TestDockerCompose(t *testing.T) {
	t.Parallel()

	// Test development docker-compose
	t.Run("validate development compose file", func(t *testing.T) {
		composeFile := "../docker-compose.dev.yml"

		// Validate compose file syntax
		cmd := shell.Command{
			Command: "docker-compose",
			Args:    []string{"-f", composeFile, "config"},
		}
		output := shell.RunCommandAndGetOutput(t, cmd)
		assert.Contains(t, output, "services:")
		assert.Contains(t, output, "backend:")
		assert.Contains(t, output, "frontend:")
		assert.Contains(t, output, "postgres:")
		assert.Contains(t, output, "redis:")

		// Validate network configuration
		assert.Contains(t, output, "api-marketplace-network")

		// Validate volume configuration
		assert.Contains(t, output, "postgres-data:")
		assert.Contains(t, output, "redis-data:")

		// Validate environment variables
		assert.Contains(t, output, "DATABASE_URL")
		assert.Contains(t, output, "REDIS_URL")
	})

	// Test production docker-compose
	t.Run("validate production compose file", func(t *testing.T) {
		composeFile := "../docker-compose.prod.yml"

		cmd := shell.Command{
			Command: "docker-compose",
			Args:    []string{"-f", composeFile, "config"},
		}
		output := shell.RunCommandAndGetOutput(t, cmd)

		// Production should have restart policies
		assert.Contains(t, output, "restart: always")

		// Production should have resource limits
		assert.Contains(t, output, "limits:")
		assert.Contains(t, output, "cpus:")
		assert.Contains(t, output, "memory:")

		// Production should have health checks
		assert.Contains(t, output, "healthcheck:")
		assert.Contains(t, output, "test:")
		assert.Contains(t, output, "interval:")
		assert.Contains(t, output, "timeout:")
		assert.Contains(t, output, "retries:")
	})
}

func TestDockerSecurity(t *testing.T) {
	t.Parallel()

	// Test security scanning
	t.Run("scan images for vulnerabilities", func(t *testing.T) {
		images := []string{
			"api-marketplace-backend:latest",
			"api-marketplace-frontend:latest",
		}

		for _, image := range images {
			// Use trivy to scan for vulnerabilities
			cmd := shell.Command{
				Command: "trivy",
				Args:    []string{"image", "--severity", "HIGH,CRITICAL", "--exit-code", "1", image},
			}

			// This should pass if no HIGH or CRITICAL vulnerabilities
			output := shell.RunCommandAndGetOutputE(t, cmd)
			if output != "" {
				t.Logf("Security scan output for %s: %s", image, output)
			}
		}
	})

	// Test Dockerfile best practices
	t.Run("validate Dockerfile security practices", func(t *testing.T) {
		dockerfiles := []string{
			"../docker/backend/Dockerfile",
			"../docker/frontend/Dockerfile",
		}

		for _, dockerfile := range dockerfiles {
			content, err := ioutil.ReadFile(dockerfile)
			require.NoError(t, err)

			contentStr := string(content)

			// Should not run as root
			assert.Contains(t, contentStr, "USER node")

			// Should use specific base image versions (not latest)
			assert.NotContains(t, contentStr, ":latest")

			// Should use COPY instead of ADD
			assert.Contains(t, contentStr, "COPY")
			assert.NotContains(t, contentStr, "ADD http")

			// Should set NODE_ENV
			assert.Contains(t, contentStr, "ENV NODE_ENV")

			// Should use multi-stage builds
			assert.Contains(t, contentStr, "FROM .* AS")
		}
	})

	// Test container runtime security
	t.Run("validate container security settings", func(t *testing.T) {
		tag := "security-test:latest"
		docker.Build(t, "../docker/backend", &docker.BuildOptions{Tags: []string{tag}})
		defer docker.DeleteImage(t, tag)

		// Test running with security options
		opts := &docker.RunOptions{
			Remove: true,
			OtherOptions: []string{
				"--cap-drop=ALL",
				"--cap-add=NET_BIND_SERVICE",
				"--read-only",
				"--security-opt=no-new-privileges",
				"--cpus=1",
				"--memory=512m",
			},
			Command: []string{"node", "--version"},
		}

		output := docker.Run(t, tag, opts)
		assert.Contains(t, output, "v") // Should output node version
	})
}

func TestDockerHealthChecks(t *testing.T) {
	t.Parallel()

	// Test health check configurations
	t.Run("validate health check endpoints", func(t *testing.T) {
		services := map[string]struct {
			port     string
			endpoint string
			expected int
		}{
			"backend": {
				port:     "3000",
				endpoint: "/health",
				expected: 200,
			},
			"frontend": {
				port:     "3001",
				endpoint: "/",
				expected: 200,
			},
		}

		// Start services with docker-compose
		composeFile := "../docker-compose.test.yml"
		projectName := fmt.Sprintf("healthcheck-test-%d", time.Now().Unix())

		// Start services
		shell.RunCommand(t, shell.Command{
			Command: "docker-compose",
			Args:    []string{"-f", composeFile, "-p", projectName, "up", "-d"},
		})

		// Cleanup
		defer shell.RunCommand(t, shell.Command{
			Command: "docker-compose",
			Args:    []string{"-f", composeFile, "-p", projectName, "down", "-v"},
		})

		// Wait for services to be healthy
		time.Sleep(10 * time.Second)

		// Test each service health endpoint
		for service, config := range services {
			t.Run(service+" health check", func(t *testing.T) {
				url := fmt.Sprintf("http://localhost:%s%s", config.port, config.endpoint)
				resp, err := http.Get(url)
				require.NoError(t, err)
				defer resp.Body.Close()

				assert.Equal(t, config.expected, resp.StatusCode)
			})
		}
	})
}

func TestDockerNetworking(t *testing.T) {
	t.Parallel()

	// Test network isolation
	t.Run("validate network isolation", func(t *testing.T) {
		// Create isolated networks
		network1 := fmt.Sprintf("test-network-1-%d", time.Now().Unix())
		network2 := fmt.Sprintf("test-network-2-%d", time.Now().Unix())

		// Create networks
		shell.RunCommand(t, shell.Command{
			Command: "docker",
			Args:    []string{"network", "create", network1},
		})
		defer shell.RunCommand(t, shell.Command{
			Command: "docker",
			Args:    []string{"network", "rm", network1},
		})

		shell.RunCommand(t, shell.Command{
			Command: "docker",
			Args:    []string{"network", "create", network2},
		})
		defer shell.RunCommand(t, shell.Command{
			Command: "docker",
			Args:    []string{"network", "rm", network2},
		})

		// Run containers in different networks
		container1 := docker.Run(t, "alpine:latest", &docker.RunOptions{
			Name:    fmt.Sprintf("container1-%d", time.Now().Unix()),
			Network: network1,
			Remove:  true,
			Detach:  true,
			Command: []string{"sleep", "30"},
		})
		defer docker.Stop(t, []string{container1}, &docker.StopOptions{})

		container2 := docker.Run(t, "alpine:latest", &docker.RunOptions{
			Name:    fmt.Sprintf("container2-%d", time.Now().Unix()),
			Network: network2,
			Remove:  true,
			Detach:  true,
			Command: []string{"sleep", "30"},
		})
		defer docker.Stop(t, []string{container2}, &docker.StopOptions{})

		// Get container IPs
		inspect1 := shell.RunCommandAndGetOutput(t, shell.Command{
			Command: "docker",
			Args:    []string{"inspect", container1, "--format", "{{.NetworkSettings.Networks." + network1 + ".IPAddress}}"},
		})

		// Try to ping from container2 to container1 (should fail)
		cmd := shell.Command{
			Command: "docker",
			Args:    []string{"exec", container2, "ping", "-c", "1", "-W", "1", inspect1},
		}
		_, err := shell.RunCommandAndGetOutputE(t, cmd)
		assert.Error(t, err) // Should fail due to network isolation
	})
}

func TestDockerVolumes(t *testing.T) {
	t.Parallel()

	// Test volume persistence
	t.Run("validate volume data persistence", func(t *testing.T) {
		volumeName := fmt.Sprintf("test-volume-%d", time.Now().Unix())

		// Create volume
		shell.RunCommand(t, shell.Command{
			Command: "docker",
			Args:    []string{"volume", "create", volumeName},
		})
		defer shell.RunCommand(t, shell.Command{
			Command: "docker",
			Args:    []string{"volume", "rm", volumeName},
		})

		// Write data to volume
		container1 := docker.Run(t, "alpine:latest", &docker.RunOptions{
			Remove: true,
			Volumes: map[string]string{
				volumeName: "/data",
			},
			Command: []string{"sh", "-c", "echo 'test data' > /data/test.txt"},
		})

		// Read data from volume with different container
		output := docker.Run(t, "alpine:latest", &docker.RunOptions{
			Remove: true,
			Volumes: map[string]string{
				volumeName: "/data",
			},
			Command: []string{"cat", "/data/test.txt"},
		})

		assert.Contains(t, output, "test data")
	})

	// Test volume backup
	t.Run("validate volume backup process", func(t *testing.T) {
		volumeName := fmt.Sprintf("backup-test-%d", time.Now().Unix())
		backupFile := fmt.Sprintf("/tmp/backup-%d.tar", time.Now().Unix())

		// Create volume with data
		shell.RunCommand(t, shell.Command{
			Command: "docker",
			Args:    []string{"volume", "create", volumeName},
		})
		defer shell.RunCommand(t, shell.Command{
			Command: "docker",
			Args:    []string{"volume", "rm", volumeName},
		})

		// Add test data
		docker.Run(t, "alpine:latest", &docker.RunOptions{
			Remove: true,
			Volumes: map[string]string{
				volumeName: "/data",
			},
			Command: []string{"sh", "-c", "echo 'important data' > /data/important.txt"},
		})

		// Backup volume
		docker.Run(t, "alpine:latest", &docker.RunOptions{
			Remove: true,
			Volumes: map[string]string{
				volumeName: "/data:ro",
				"/tmp":     "/backup",
			},
			Command: []string{"tar", "cf", fmt.Sprintf("/backup/backup-%d.tar", time.Now().Unix()), "-C", "/data", "."},
		})

		// Verify backup exists
		_, err := shell.RunCommandAndGetOutputE(t, shell.Command{
			Command: "ls",
			Args:    []string{backupFile},
		})
		assert.NoError(t, err)
	})
}

// Helper function to wait for container to be ready
func waitForContainer(t *testing.T, containerId string, maxRetries int) {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	require.NoError(t, err)

	for i := 0; i < maxRetries; i++ {
		container, err := cli.ContainerInspect(context.Background(), containerId)
		require.NoError(t, err)

		if container.State.Running && container.State.Health.Status == "healthy" {
			return
		}

		time.Sleep(2 * time.Second)
	}

	t.Fatalf("Container %s did not become ready in time", containerId)
}