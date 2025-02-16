package e2e

import (
	"log"
	"os"
	"os/exec"
	"testing"
	"time"
)

const baseURL = "http://localhost:8081/api/"

func TestMain(m *testing.M) {
	startDocker()
	time.Sleep(30 * time.Second)
	exitCode := m.Run()
	stopDocker()
	os.Exit(exitCode)
}

func startDocker() {
	log.Println("Starting docker-compose...")
	cmd := exec.Command("docker", "compose", "-f", "../../.docker-compose.test.yaml", "up", "-d")
	if err := cmd.Run(); err != nil {
		log.Fatalf("Could not start docker-compose: %v", err)
	}
}

func stopDocker() {
	log.Println("Stopping docker-compose...")
	cmd := exec.Command("docker", "compose", "-f", "../../.docker-compose.test.yaml", "down")
	if err := cmd.Run(); err != nil {
		log.Fatalf("Could not stop docker-compose: %v", err)
	}
}
