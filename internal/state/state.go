package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// State represents the persistent state of notifications
type State struct {
	Entries map[string]map[int]time.Time `json:"entries"` // domain -> threshold -> last sent time
}

// Manager handles state persistence
type Manager struct {
	filePath      string
	cooldownHours int
	state         *State
}

// NewManager creates a new state manager
func NewManager(filePath string, cooldownHours int) (*Manager, error) {
	m := &Manager{
		filePath:      filePath,
		cooldownHours: cooldownHours,
		state: &State{
			Entries: make(map[string]map[int]time.Time),
		},
	}

	if err := m.load(); err != nil {
		return nil, fmt.Errorf("failed to load state: %w", err)
	}

	return m, nil
}

// load reads the state from disk
func (m *Manager) load() error {
	if m.filePath == "" {
		return nil
	}

	data, err := os.ReadFile(m.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to read state file: %w", err)
	}

	if err := json.Unmarshal(data, &m.state); err != nil {
		return fmt.Errorf("failed to unmarshal state: %w", err)
	}

	// Ensure nested maps exist
	for domain, thresholds := range m.state.Entries {
		if thresholds == nil {
			m.state.Entries[domain] = make(map[int]time.Time)
		}
	}

	return nil
}

// save writes the state to disk
func (m *Manager) save() error {
	if m.filePath == "" {
		return nil
	}

	// Ensure directory exists
	dir := filepath.Dir(m.filePath)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			// If we can't create the directory, try to save to current directory instead
			// This handles cases where user doesn't have permission to create /var/lib/ssl-monitor
			fallbackFile := filepath.Base(m.filePath)
			if fallbackFile == "" {
				fallbackFile = "ssl-monitor-state.json"
			}
			m.filePath = fallbackFile
		}
	}

	data, err := json.MarshalIndent(m.state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	// Try to write the file
	if err := os.WriteFile(m.filePath, data, 0644); err != nil {
		// If that fails, try to write to a file in the current directory
		fallbackFile := "ssl-monitor-state.json"
		if err := os.WriteFile(fallbackFile, data, 0644); err != nil {
			return fmt.Errorf("failed to write state file: %w", err)
		}
		m.filePath = fallbackFile
	}

	return nil
}

// ShouldSend checks if a notification should be sent for a domain and threshold
func (m *Manager) ShouldSend(domain string, threshold int) bool {
	domainMap, exists := m.state.Entries[domain]
	if !exists {
		return true
	}

	lastSent, exists := domainMap[threshold]
	if !exists {
		return true
	}

	cooldown := time.Duration(m.cooldownHours) * time.Hour
	return time.Since(lastSent) > cooldown
}

// MarkSent records that a notification was sent for a domain and threshold
func (m *Manager) MarkSent(domain string, threshold int) error {
	if _, exists := m.state.Entries[domain]; !exists {
		m.state.Entries[domain] = make(map[int]time.Time)
	}
	m.state.Entries[domain][threshold] = time.Now()
	return m.save()
}

// Clear removes all state entries
func (m *Manager) Clear() error {
	m.state.Entries = make(map[string]map[int]time.Time)
	return m.save()
}

// GetLastSent returns the last sent time for a domain and threshold
func (m *Manager) GetLastSent(domain string, threshold int) (time.Time, bool) {
	domainMap, exists := m.state.Entries[domain]
	if !exists {
		return time.Time{}, false
	}
	lastSent, exists := domainMap[threshold]
	return lastSent, exists
}