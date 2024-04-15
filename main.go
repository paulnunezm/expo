package main

import (
	"fmt"
	"log"
	"strconv"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	expected int
	actual   int

	expectedInput textinput.Model
}

func initialModel() *model {
	inputField := textinput.New()
	inputField.Placeholder = "Enter expected pomodoros"
	inputField.Focus()

	return &model{
		expected:      0,
		actual:        0,
		expectedInput: inputField,
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return nil, tea.Quit
		case "enter":
			num, err := strconv.Atoi(m.expectedInput.Value())
			if err != nil {
				log.Printf("error converting %s", m.expectedInput.Value())
				return m, tea.Quit
			}
			m.expected = num
			m.expectedInput.Blur()

			return m, nil
		}
	}
	m.expectedInput, cmd = m.expectedInput.Update(msg)

	return m, cmd
}

func (m model) View() string {
	if m.expected == 0 {
		return m.expectedInput.View()
	} else {
		return fmt.Sprintf("Value entered -> %d", m.expected)
	}
}

func main() {

	m := initialModel()
	f, err := tea.LogToFile("debug.log", "debug")
	if err != nil {
		log.Fatalf("err: %w", err)
	}
	defer f.Close()
	//p := tea.NewProgram(m, tea.WithAltScreen())
	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}

}
