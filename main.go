package main

import (
	"fmt"
	"log"
	"strconv"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	expected         int
	actual           int
	canStartPomodoro bool

	expectedInput textinput.Model
}

func initialModel() *model {
	inputField := textinput.New()
	inputField.Placeholder = "Enter expected pomodoros"
	inputField.Focus()

	return &model{
		expected:         0,
		actual:           0,
		expectedInput:    inputField,
		canStartPomodoro: false,
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
			return m, tea.Quit

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
	s := "\n"
	if m.expected == 0 {
		s += "======================================================\n"
		s += "=====|| Enter the number of expected pomodoros ||=====\n"
		s += "======================================================\n\n"

		s += m.expectedInput.View()
		s += "\n\n======================================================\n"
		return s
	} else {
		s += "======================================================\n"
		s += " - Expected pomodoros: %d\n"
		s += " - Actual pomodoros: %d\n"
		s += "\n\n\n\nPress <enter> to start pomodoro <q> to quit"
		s += "\n======================================================"
		return fmt.Sprintf(s, m.expected, m.actual)
	}
}

func main() {

	m := initialModel()
	f, err := tea.LogToFile("debug.log", "debug")
	if err != nil {
		log.Fatalf(`err:%w`, err)
	}
	defer f.Close()
	p := tea.NewProgram(m, tea.WithAltScreen())
	// p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}

}
