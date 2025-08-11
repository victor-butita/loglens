package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// --- Step 1: Define the Model ---
type model struct {
	logEntries []map[string]interface{}
}

// Init is the first function that will be called. It returns an initial command.
func (m model) Init() tea.Cmd {
	return watchLogFile()
}

// --- Step 2: Define Messages ---
type allLogsReadMsg struct {
	entries []map[string]interface{}
}

// This command reads the entire file and returns one single message with all the data.
func watchLogFile() tea.Cmd {
	return func() tea.Msg {
		file, err := os.Open("logs.jsonl")
		if err != nil {
			fmt.Println("Error opening file:", err)
			return nil
		}
		defer file.Close()

		var allEntries []map[string]interface{}
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			var logEntry map[string]interface{}
			if err := json.Unmarshal([]byte(line), &logEntry); err == nil {
				allEntries = append(allEntries, logEntry)
			}
		}
		// Let's add a delay to simulate a slow file read.
		time.Sleep(1 * time.Second)

		// Return one single message containing all the log entries.
		return allLogsReadMsg{entries: allEntries}
	}
}

// --- Step 3: Define the Update function ---
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

	case allLogsReadMsg:
		m.logEntries = msg.entries
		return m, nil
	}

	return m, nil
}

// --- Step 4: Define the View function ---
func (m model) View() string {
	s := "LogLens - Reading logs...\n\n"
	if len(m.logEntries) > 0 {
		s = "LogLens - Logs loaded!\n\n"
	}
	s += fmt.Sprintf("%d log entries loaded.\n\n", len(m.logEntries))
	s += "Press 'q' to quit.\n"
	return s
}

// --- Step 5: Run the application ---
func main() {
	// Create a new Bubble Tea program. We pass it the initial state of our model.
	// The program will automatically call the Init method on our model.
	p := tea.NewProgram(model{logEntries: []map[string]interface{}{}}, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}