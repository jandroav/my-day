package llm

import (
	"context"
	"fmt"
	"net/http"
	"os/exec"
	"strings"
	"time"
	
	"github.com/fatih/color"
)

// DockerLLMManager handles automatic Docker container management for LLM
type DockerLLMManager struct {
	containerName string
	imageName     string
	port          string
	baseURL       string
	model         string
}

// NewDockerLLMManager creates a new Docker LLM manager
func NewDockerLLMManager() *DockerLLMManager {
	return &DockerLLMManager{
		containerName: "my-day-ollama",
		imageName:     "ollama/ollama",
		port:          "11434",
		baseURL:       "http://localhost:11434",
		model:         "qwen2.5:3b", // Fast, high-quality model optimized for summarization
	}
}

// IsDockerAvailable checks if Docker is installed and running
func (d *DockerLLMManager) IsDockerAvailable() bool {
	cmd := exec.Command("docker", "ps")
	return cmd.Run() == nil
}

// IsContainerRunning checks if the LLM container is already running
func (d *DockerLLMManager) IsContainerRunning() bool {
	cmd := exec.Command("docker", "ps", "--filter", fmt.Sprintf("name=%s", d.containerName), "--format", "{{.Names}}")
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.Contains(string(output), d.containerName)
}

// IsModelLoaded checks if the model is loaded in Ollama
func (d *DockerLLMManager) IsModelLoaded() bool {
	resp, err := http.Get(d.baseURL + "/api/tags")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == 200
}

// StartContainer starts the Ollama Docker container
func (d *DockerLLMManager) StartContainer() error {
	color.Cyan("🐳 Starting Docker LLM container...")
	
	// Check if container exists but is stopped
	if d.containerExists() {
		color.White("🔄 Starting existing container...")
		cmd := exec.Command("docker", "start", d.containerName)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to start existing container: %w", err)
		}
	} else {
		// Create and run new container
		color.White("📦 Creating new LLM container...")
		cmd := exec.Command("docker", "run", "-d",
			"--name", d.containerName,
			"-p", d.port+":11434",
			"-v", "my-day-ollama:/root/.ollama",
			d.imageName)
		
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to create container: %w", err)
		}
	}
	
	// Wait for container to be ready
	color.White("⏳ Waiting for container to be ready...")
	return d.waitForContainer()
}

// containerExists checks if the container exists (running or stopped)
func (d *DockerLLMManager) containerExists() bool {
	cmd := exec.Command("docker", "ps", "-a", "--filter", fmt.Sprintf("name=%s", d.containerName), "--format", "{{.Names}}")
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.Contains(string(output), d.containerName)
}

// waitForContainer waits for the container to be ready
func (d *DockerLLMManager) waitForContainer() error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for container to be ready")
		default:
			resp, err := http.Get(d.baseURL)
			if err == nil {
				resp.Body.Close()
				if resp.StatusCode == 200 {
					color.Green("✅ Container ready!")
					return nil
				}
			}
			time.Sleep(2 * time.Second)
		}
	}
}

// PullModel downloads the LLM model if not present
func (d *DockerLLMManager) PullModel() error {
	color.Cyan("🧠 Setting up LLM model...")
	
	// Check if model is already available
	if d.IsModelLoaded() {
		cmd := exec.Command("docker", "exec", d.containerName, "ollama", "list")
		output, err := cmd.Output()
		if err == nil && strings.Contains(string(output), d.model) {
			color.Green("✅ Model already available!")
			return nil
		}
	}
	
	color.White("📥 Downloading LLM model (this may take a few minutes on first run)...")
	cmd := exec.Command("docker", "exec", d.containerName, "ollama", "pull", d.model)
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to pull model: %w", err)
	}
	
	color.Green("✅ Model ready!")
	return nil
}

// EnsureReady ensures the Docker LLM is ready for use
func (d *DockerLLMManager) EnsureReady() error {
	if !d.IsDockerAvailable() {
		return fmt.Errorf("Docker is required for LLM functionality. Please install and start Docker")
	}
	
	if !d.IsContainerRunning() {
		if err := d.StartContainer(); err != nil {
			return err
		}
	}
	
	return d.PullModel()
}

// GetBaseURL returns the LLM base URL
func (d *DockerLLMManager) GetBaseURL() string {
	return d.baseURL
}

// GetModel returns the model name
func (d *DockerLLMManager) GetModel() string {
	return d.model
}

// StopContainer stops the LLM container
func (d *DockerLLMManager) StopContainer() error {
	if !d.IsContainerRunning() {
		return nil
	}
	
	color.Yellow("🛑 Stopping LLM container...")
	cmd := exec.Command("docker", "stop", d.containerName)
	return cmd.Run()
}

// GetStatus returns the current status of the Docker LLM
func (d *DockerLLMManager) GetStatus() string {
	if !d.IsDockerAvailable() {
		return "❌ Docker not available"
	}
	
	if !d.IsContainerRunning() {
		return "⏹️  Container stopped"
	}
	
	if !d.IsModelLoaded() {
		return "⏳ Model loading"
	}
	
	return "✅ Ready"
}