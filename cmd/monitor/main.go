package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/fatih/color"
)

// LogEntry represents a structured log entry
type LogEntry struct {
	Time     string `json:"time"`
	Level    string `json:"level"`
	Message  string `json:"msg"`
	Service  string `json:"-"`
	RawLine  string `json:"-"`
	Method   string `json:"method,omitempty"`
	URI      string `json:"uri,omitempty"`
	Status   int    `json:"status,omitempty"`
	Latency  string `json:"latency,omitempty"`
	RemoteIP string `json:"remote_ip,omitempty"`
	Error    string `json:"error,omitempty"`
}

// LogMonitor handles live monitoring of all services
type LogMonitor struct {
	ctx      context.Context
	cancel   context.CancelFunc
	wg       sync.WaitGroup
	colorMap map[string]*color.Color
	services []string
}

// NewLogMonitor creates a new log monitor instance
func NewLogMonitor() *LogMonitor {
	ctx, cancel := context.WithCancel(context.Background())

	return &LogMonitor{
		ctx:    ctx,
		cancel: cancel,
		colorMap: map[string]*color.Color{
			"app":      color.New(color.FgCyan, color.Bold),
			"postgres": color.New(color.FgGreen, color.Bold),
			"caddy":    color.New(color.FgYellow, color.Bold),
			"system":   color.New(color.FgMagenta, color.Bold),
		},
		services: []string{"app", "postgres", "caddy"},
	}
}

// Start begins monitoring all services
func (m *LogMonitor) Start() error {
	// Print header
	m.printHeader()

	// Start Docker Compose logs
	m.wg.Add(1)
	go m.monitorDockerLogs()

	// Start health check monitor
	m.wg.Add(1)
	go m.monitorHealth()

	// Start metrics monitor (if available)
	m.wg.Add(1)
	go m.monitorMetrics()

	// Handle shutdown gracefully
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		fmt.Println("\n" + m.colorMap["system"].Sprint("🛑 Shutting down monitor..."))
		m.cancel()
	}()

	m.wg.Wait()
	return nil
}

// printHeader displays the monitor startup information
func (m *LogMonitor) printHeader() {
	header := color.New(color.FgWhite, color.Bold, color.BgBlue)
	separator := strings.Repeat("=", 80)

	fmt.Println(separator)
	header.Println("🚀 GO WEB SERVER - LIVE LOG MONITOR")
	fmt.Printf("📅 Started: %s\n", time.Now().Format("2006-01-02 15:04:05 MST"))
	fmt.Printf("🐳 Monitoring: %s\n", strings.Join(m.services, ", "))
	fmt.Println("💡 Press Ctrl+C to stop monitoring")
	fmt.Println(separator)
	fmt.Println()
}

// monitorDockerLogs tails Docker Compose logs
func (m *LogMonitor) monitorDockerLogs() {
	defer m.wg.Done()

	cmd := exec.CommandContext(m.ctx, "docker", "compose", "logs", "-f", "--no-log-prefix", "-t")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		m.logError("Failed to create docker logs pipe", err)
		return
	}

	if err := cmd.Start(); err != nil {
		m.logError("Failed to start docker compose logs", err)
		return
	}

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		select {
		case <-m.ctx.Done():
			return
		default:
			line := scanner.Text()
			m.processLogLine(line)
		}
	}

	if err := scanner.Err(); err != nil {
		m.logError("Error reading docker logs", err)
	}
}

// processLogLine parses and formats a log line
func (m *LogMonitor) processLogLine(line string) {
	if line == "" {
		return
	}

	// Parse Docker Compose log format: timestamp service | message
	parts := strings.SplitN(line, " ", 3)
	if len(parts) < 3 {
		return
	}

	timestamp := parts[0]
	serviceFull := parts[1]
	message := parts[2]

	// Extract service name from "gowebserver-app" -> "app"
	service := m.extractServiceName(serviceFull)

	// Create log entry
	entry := LogEntry{
		Time:    timestamp,
		Service: service,
		RawLine: message,
	}

	// Try to parse as JSON for structured logs
	if strings.HasPrefix(message, "{") {
		var jsonEntry map[string]interface{}
		if err := json.Unmarshal([]byte(message), &jsonEntry); err == nil {
			m.parseJSONLog(&entry, jsonEntry)
		}
	} else {
		// Parse Go app logs (key=value format)
		m.parseGoAppLog(&entry, message)
	}

	// Display the formatted log
	m.displayLog(entry)
}

// extractServiceName extracts service name from Docker container name
func (m *LogMonitor) extractServiceName(containerName string) string {
	// Remove "gowebserver-" prefix and any trailing numbers
	name := strings.TrimPrefix(containerName, "gowebserver-")
	name = strings.TrimSuffix(name, "|")
	return strings.TrimSpace(name)
}

// parseJSONLog parses JSON structured logs (Caddy)
func (m *LogMonitor) parseJSONLog(entry *LogEntry, jsonData map[string]interface{}) {
	if level, ok := jsonData["level"].(string); ok {
		entry.Level = level
	}
	if msg, ok := jsonData["msg"].(string); ok {
		entry.Message = msg
	}
	if ts, ok := jsonData["ts"].(float64); ok {
		entry.Time = time.Unix(int64(ts), 0).Format("15:04:05")
	}

	// Parse HTTP access logs
	if request, ok := jsonData["request"].(map[string]interface{}); ok {
		if method, ok := request["method"].(string); ok {
			entry.Method = method
		}
		if uri, ok := request["uri"].(string); ok {
			entry.URI = uri
		}
		if remoteIP, ok := request["remote_ip"].(string); ok {
			entry.RemoteIP = remoteIP
		}
	}

	if status, ok := jsonData["status"].(float64); ok {
		entry.Status = int(status)
	}
	if duration, ok := jsonData["duration"].(float64); ok {
		entry.Latency = fmt.Sprintf("%.2fms", duration*1000)
	}
}

// parseGoAppLog parses Go application logs (key=value format)
func (m *LogMonitor) parseGoAppLog(entry *LogEntry, message string) {
	// Parse time=... level=... msg=... format
	fields := m.parseKeyValuePairs(message)

	if timeStr, ok := fields["time"]; ok {
		if t, err := time.Parse("2006-01-02T15:04:05.999Z", timeStr); err == nil {
			entry.Time = t.Format("15:04:05.999")
		}
	}

	entry.Level = fields["level"]
	entry.Message = fields["msg"]
	entry.Method = fields["method"]
	entry.URI = fields["uri"]
	entry.Latency = fields["latency"]
	entry.RemoteIP = fields["remote_ip"]
	entry.Error = fields["error"]

	if status := fields["status"]; status != "" {
		fmt.Sscanf(status, "%d", &entry.Status)
	}
}

// parseKeyValuePairs parses key=value pairs from log messages
func (m *LogMonitor) parseKeyValuePairs(message string) map[string]string {
	fields := make(map[string]string)
	parts := strings.Fields(message)

	for _, part := range parts {
		if kv := strings.SplitN(part, "=", 2); len(kv) == 2 {
			key := kv[0]
			value := strings.Trim(kv[1], `"`)
			fields[key] = value
		}
	}

	return fields
}

// displayLog formats and displays a log entry
func (m *LogMonitor) displayLog(entry LogEntry) {
	serviceColor := m.colorMap[entry.Service]
	if serviceColor == nil {
		serviceColor = color.New(color.FgWhite)
	}

	// Format timestamp
	timeColor := color.New(color.FgBlue)
	timestamp := timeColor.Sprintf("[%s]", entry.Time)

	// Format service name
	serviceName := serviceColor.Sprintf("%-8s", strings.ToUpper(entry.Service))

	// Format level with color
	levelStr := m.formatLevel(entry.Level)

	// Build main message
	var message string
	if entry.Method != "" && entry.URI != "" {
		// HTTP request log
		statusColor := m.getStatusColor(entry.Status)
		message = fmt.Sprintf("%s %s → %s %s",
			color.New(color.FgWhite, color.Bold).Sprint(entry.Method),
			entry.URI,
			statusColor.Sprintf("%d", entry.Status),
			color.New(color.FgHiBlack).Sprintf("(%s)", entry.Latency))

		if entry.RemoteIP != "" {
			message += color.New(color.FgHiBlack).Sprintf(" from %s", entry.RemoteIP)
		}
	} else {
		message = entry.Message
		if entry.Error != "" {
			message += " " + color.New(color.FgRed).Sprintf("error=%s", entry.Error)
		}
	}

	// Print the formatted log line
	fmt.Printf("%s %s %s %s\n", timestamp, serviceName, levelStr, message)
}

// formatLevel returns a colored level indicator
func (m *LogMonitor) formatLevel(level string) string {
	switch strings.ToUpper(level) {
	case "ERROR", "FATAL":
		return color.New(color.FgRed, color.Bold).Sprintf("[%-5s]", "ERROR")
	case "WARN", "WARNING":
		return color.New(color.FgYellow, color.Bold).Sprintf("[%-5s]", "WARN")
	case "INFO":
		return color.New(color.FgGreen).Sprintf("[%-5s]", "INFO")
	case "DEBUG":
		return color.New(color.FgHiBlack).Sprintf("[%-5s]", "DEBUG")
	default:
		return color.New(color.FgWhite).Sprintf("[%-5s]", strings.ToUpper(level))
	}
}

// getStatusColor returns appropriate color for HTTP status codes
func (m *LogMonitor) getStatusColor(status int) *color.Color {
	switch {
	case status >= 200 && status < 300:
		return color.New(color.FgGreen, color.Bold)
	case status >= 300 && status < 400:
		return color.New(color.FgCyan, color.Bold)
	case status >= 400 && status < 500:
		return color.New(color.FgYellow, color.Bold)
	case status >= 500:
		return color.New(color.FgRed, color.Bold)
	default:
		return color.New(color.FgWhite)
	}
}

// monitorHealth periodically checks service health
func (m *LogMonitor) monitorHealth() {
	defer m.wg.Done()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.checkHealth()
		}
	}
}

// checkHealth performs health checks and displays status
func (m *LogMonitor) checkHealth() {
	timestamp := color.New(color.FgBlue).Sprintf("[%s]", time.Now().Format("15:04:05"))
	serviceName := m.colorMap["system"].Sprintf("%-8s", "HEALTH")

	// Check application health
	cmd := exec.Command("curl", "-s", "-f", "http://localhost:8080/health")
	if err := cmd.Run(); err != nil {
		level := color.New(color.FgRed, color.Bold).Sprint("[ERROR]")
		message := "Application health check failed"
		fmt.Printf("%s %s %s %s\n", timestamp, serviceName, level, message)
	} else {
		level := color.New(color.FgGreen).Sprint("[INFO ]")
		message := "✓ Application is healthy"
		fmt.Printf("%s %s %s %s\n", timestamp, serviceName, level, message)
	}
}

// monitorMetrics periodically checks metrics endpoint
func (m *LogMonitor) monitorMetrics() {
	defer m.wg.Done()

	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.checkMetrics()
		}
	}
}

// checkMetrics checks if metrics endpoint is available
func (m *LogMonitor) checkMetrics() {
	timestamp := color.New(color.FgBlue).Sprintf("[%s]", time.Now().Format("15:04:05"))
	serviceName := m.colorMap["system"].Sprintf("%-8s", "METRICS")

	cmd := exec.Command("curl", "-s", "-f", "http://localhost:8080/metrics")
	if err := cmd.Run(); err != nil {
		level := color.New(color.FgYellow, color.Bold).Sprint("[WARN ]")
		message := "Metrics endpoint not available"
		fmt.Printf("%s %s %s %s\n", timestamp, serviceName, level, message)
	} else {
		level := color.New(color.FgGreen).Sprint("[INFO ]")
		message := "✓ Metrics endpoint is available"
		fmt.Printf("%s %s %s %s\n", timestamp, serviceName, level, message)
	}
}

// logError displays an error message
func (m *LogMonitor) logError(message string, err error) {
	timestamp := color.New(color.FgBlue).Sprintf("[%s]", time.Now().Format("15:04:05"))
	serviceName := m.colorMap["system"].Sprintf("%-8s", "MONITOR")
	level := color.New(color.FgRed, color.Bold).Sprint("[ERROR]")
	errorMsg := fmt.Sprintf("%s: %v", message, err)

	fmt.Printf("%s %s %s %s\n", timestamp, serviceName, level, errorMsg)
}

func main() {
	monitor := NewLogMonitor()

	fmt.Println("🔍 Starting Go Web Server Live Log Monitor...")
	fmt.Println("📊 Monitoring: Application, PostgreSQL, Caddy, Health & Metrics")
	fmt.Println()

	if err := monitor.Start(); err != nil {
		log.Fatalf("Monitor failed: %v", err)
	}

	fmt.Println("👋 Monitor stopped.")
}
