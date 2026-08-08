package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const (
	proxyStarted = "start proxy success"
	startTimeout = 30 * time.Second
)

const configTemplate = `serverAddr = "%s"
serverPort = 443
transport.tls.enable = true
transport.tls.serverName = "%s"
transport.tls.disableCustomTLSFirstByte = true
metadatas.token = "%s"

[[proxies]]
name = "%s-smtp"
type = "tcpmux"
multiplexer = "httpconnect"
customDomains = ["%s"]
localIP = "127.0.0.1"
localPort = %d
`

type Tunnels struct {
	frpc       string
	serverAddr string
	serverName string
	localPort  int
	workDir    string
	mutex      sync.Mutex
	running    map[string]*exec.Cmd
}

func NewTunnels(frpc string, serverAddr string, serverName string, localPort int,
	workDir string) *Tunnels {
	return &Tunnels{
		frpc: frpc, serverAddr: serverAddr, serverName: serverName,
		localPort: localPort, workDir: workDir, running: map[string]*exec.Cmd{},
	}
}

func (t *Tunnels) Start(domain string, token string) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	if _, found := t.running[domain]; found {
		return fmt.Errorf("tunnel for %s is already running", domain)
	}
	path := filepath.Join(t.workDir, fmt.Sprintf("frpc-%s.toml", domain))
	config := fmt.Sprintf(configTemplate, t.serverAddr, t.serverName, token, domain,
		domain, t.localPort)
	if err := os.WriteFile(path, []byte(config), 0644); err != nil {
		return err
	}
	command := exec.Command(t.frpc, "-c", path)
	output, err := command.StdoutPipe()
	if err != nil {
		return err
	}
	command.Stderr = command.Stdout
	if err := command.Start(); err != nil {
		return err
	}
	started := make(chan struct{})
	go func() {
		logFile, _ := os.Create(filepath.Join(t.workDir, fmt.Sprintf("frpc-%s.log", domain)))
		if logFile != nil {
			defer logFile.Close()
		}
		announced := false
		scanner := bufio.NewScanner(output)
		for scanner.Scan() {
			line := scanner.Text()
			if logFile != nil {
				_, _ = fmt.Fprintln(logFile, line)
			}
			if !announced && strings.Contains(line, proxyStarted) {
				announced = true
				close(started)
			}
		}
	}()
	select {
	case <-started:
		t.running[domain] = command
		return nil
	case <-time.After(startTimeout):
		_ = command.Process.Kill()
		_ = command.Wait()
		return fmt.Errorf("tunnel for %s did not come up: %s", domain, t.log(domain))
	}
}

func (t *Tunnels) StopAll() {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	for domain, command := range t.running {
		_ = command.Process.Kill()
		_ = command.Wait()
		delete(t.running, domain)
	}
}

func (t *Tunnels) log(domain string) string {
	content, err := os.ReadFile(filepath.Join(t.workDir, fmt.Sprintf("frpc-%s.log", domain)))
	if err != nil {
		return err.Error()
	}
	return string(content)
}
