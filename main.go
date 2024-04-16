package main

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/0xAX/notificator"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/timer"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/deckarep/gosx-notifier"
)

const pomodoroMinutes = 25

// const timeout = time.Minute * 2 //pomodoroMinutes
const timeout = time.Second * 2

var notify *notificator.Notificator

type model struct {
	expected int
	actual   int
	timer    timer.Model

	canStart   bool
	hasStarted bool

	expectedInput textinput.Model
}

func initModel() *model {
	inputField := textinput.New()
	inputField.Placeholder = "Enter expected pomodoros"
	inputField.Focus()

	return &model{
		expected:      0,
		actual:        0,
		timer:         timer.NewWithInterval(timeout, time.Second),
		expectedInput: inputField,
		canStart:      false,
		hasStarted:    false,
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {

	case timer.TickMsg:
		log.Print("Timer ticking")
		m.timer, cmd = m.timer.Update(msg)
		return m, cmd

	case timer.StartStopMsg:
		log.Print("timer starting")
		m.timer, cmd = m.timer.Update(msg)
		log.Print("timer started")
		return m, cmd

	case timer.TimeoutMsg:
		m.actual++
		m.canStart = true
		m.timer.Timeout = timeout
		cmd = m.timer.Stop()
		notifyStopped()
		return m, cmd

	case tea.KeyMsg:
		switch msg.String() {

		case "ctrl+c", "q":
			return m, tea.Quit

		case "enter":
			log.Printf("enter clicked/ can start = %t", m.canStart)
			if m.canStart {
				log.Print("Trying to start/pause the timmer")
				if m.hasStarted {
					return m, m.timer.Toggle()
				} else {
					m.hasStarted = true
					return m, m.timer.Init()
				}
			} else {
				log.Print("handle text input")
				num, err := strconv.Atoi(m.expectedInput.Value())
				if err != nil {
					log.Printf("error converting %s", m.expectedInput.Value())
					return m, tea.Quit
				}
				m.canStart = true
				m.expected = num
				m.expectedInput.Blur()
				return m, nil
			}
		}
	}
	m.expectedInput, cmd = m.expectedInput.Update(msg)

	return m, cmd
}

func (m model) View() string {
	s := "\n"

	h := "\n\n\n\nPress:\n" +
		" <enter> to start or pause\n" +
		" <s> to stop\n" +
		" <q> or <crtl-c> to quit"
	if m.hasStarted && m.timer.Running() {
		s += "========================RUNNING==============================\n"
		s += " - Expected pomodoros: %d\n"
		s += " - Actual pomodoros: %d\n"
		s += m.timer.View()
		s += "\n\n======================================================"
		s += h
		return fmt.Sprintf(s, m.expected, m.actual)
	} else if m.hasStarted && !m.timer.Running() {
		s += "========================PAUSED==============================\n"
		s += " - Expected pomodoros: %d\n"
		s += " - Actual pomodoros: %d\n"
		s += m.timer.View()
		s += "\n\n======================================================"
		s += h
		return fmt.Sprintf(s, m.expected, m.actual)
	} else if m.expected == 0 {
		s += "======================================================\n"
		s += "=====|| Enter the number of expected pomodoros ||=====\n"
		s += "======================================================\n\n"

		s += m.expectedInput.View()
		s += "\n\n======================================================\n"
		s += "\n\n"
		s += h
		return s
	} else {
		s += "======================================================\n"
		s += " - Expected pomodoros: %d\n"
		s += " - Actual pomodoros: %d\n"
		s += "\n\n======================================================"
		s += h
		return fmt.Sprintf(s, m.expected, m.actual)
	}
}

func main() {
	notify = notificator.New(notificator.Options{
		DefaultIcon: "icon/default.png",
		AppName:     "ExPom",
	})

	f, err := tea.LogToFile("debug.log", "debug")
	if err != nil {
		log.Fatalf(`err:%w`, err)
	}
	defer f.Close()

	p := tea.NewProgram(initModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

func notifyStopped() {

	// THis is global
	// notify.Push("OOO", "OOO", "", notificator.UR_NORMAL)


	note := gosxnotifier.NewNotification("Check your Apple Stock!")

	note.Title = "ExPomo"
	note.Message = "🍅 Finished 🍅"
	note.Sound = gosxnotifier.Funk
	note.Sound = gosxnotifier.Hero 
	note.Group = "com.expomo"
	note.Sender = "com.apple.Safari" //Optionally, set a sender (Notification will now use the Safari icon)
	note.AppIcon = "gopher.png"      //Optionally, an app icon (10.9+ ONLY)
	note.ContentImage = "gopher.png" //Optionally, a content image (10.9+ ONLY)

	err := note.Push()
	if err != nil {
		log.Println("Uh oh!")
	}
}
