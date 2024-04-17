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
	gosxnotifier "github.com/deckarep/gosx-notifier"
)

const (
	pomodoroMinutes  = 25
	breakMinutes     = 5
	longBreakMinutes = 20
)

var notify *notificator.Notificator

type ScreenState int

const (
	Initial ScreenState = iota + 1
	Entering
	Running
	Paused
	Stopped
)

type PomodoroType int

const (
	WorkPomodoro PomodoroType = iota
	BreakPomodoro
)

type model struct {
	screenState       ScreenState
	expected          int
	actual            int
	pomodoroTimerType PomodoroType
	timer             timer.Model
	textFieldInput    textinput.Model
}

func initModel() *model {
	return &model{
		screenState:       Initial,
		expected:          0,
		actual:            0,
		pomodoroTimerType: WorkPomodoro,
		timer:             timer.NewWithInterval(time.Second*pomodoroMinutes, time.Second),
		textFieldInput:    textinput.New(),
	}
}

func (m model) Init() tea.Cmd {
	return nil // textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {

	case timer.TickMsg:
		m.timer, cmd = m.timer.Update(msg)
		return m, cmd

	case timer.StartStopMsg:
		m.timer, cmd = m.timer.Update(msg)
		return m, cmd

	case timer.TimeoutMsg:
		return handleTimerStoppedCmd(m)

	case tea.KeyMsg:
		switch msg.String() {

		case "ctrl+c", "q":
			return m, tea.Quit

		case "s", "S":
			return handleStartCmd(m)

		case "e", "E":
			m.screenState = Entering
			m.textFieldInput.Focus()
			return m, m.textFieldInput.Cursor.BlinkCmd()

		case "b", "B":
			return m, nil

		case "p", "P":
			return m, nil

		case "r", "R":
			return m, nil

		case "enter":
			if m.screenState == Entering {
				num, err := strconv.Atoi(m.textFieldInput.Value())
				if err != nil {
					log.Printf("error converting %s", m.textFieldInput.Value())
					return m, tea.Quit
				}
				m.expected = num

				m.textFieldInput.Blur()
				m.textFieldInput.Reset()
				return handleStartCmd(m)
			} else {
				return m, nil
			}
		}
	}
	m.textFieldInput, cmd = m.textFieldInput.Update(msg)

	return m, cmd
}

func handleTimerStoppedCmd(m model) (tea.Model, tea.Cmd) {
	if m.pomodoroTimerType == WorkPomodoro {
		m.actual++
	}
	m.screenState = Stopped

	timeout := time.Second * pomodoroMinutes
	if m.pomodoroTimerType == WorkPomodoro {
		m.pomodoroTimerType = BreakPomodoro
		timeout = time.Second * breakMinutes
	} else {
		m.pomodoroTimerType = WorkPomodoro
	}

	m.timer.Timeout = timeout // Resets the timer
	notifyStopped()
	return m, m.timer.Stop()
}

var wasTimerInitiated = false

func handleStartCmd(m model) (tea.Model, tea.Cmd) {
	switch m.screenState {
	case Initial:
		m.pomodoroTimerType = WorkPomodoro
		m.screenState = Running
		wasTimerInitiated = true
		return m, m.timer.Init()

	case Entering:
		m.pomodoroTimerType = WorkPomodoro
		m.screenState = Running
		if wasTimerInitiated {
			wasTimerInitiated = true
			return m, m.timer.Toggle()
		}
		return m, m.timer.Init()

	case Stopped:
		return m, m.timer.Toggle()
	}

	return m, nil
}

func (m model) View() string {
	s := ""
	switch m.screenState {
	case Initial:
		s += renderInitialState(m)
	case Paused:
		s += renderPausedState(m)
	case Running:
		s += renderRunningState(m)
	case Stopped:
		s += renderStoppedState(m)
	case Entering:
		s += renderEnteringState(m)
	}
	return s
}

func getTitle(p PomodoroType) string {
	title := ""
	title += "---------------\n"
	if p == WorkPomodoro {
		title += "Pomodoro Timer"
	} else {
		title += "Take a Break"
	}
	title += "\n---------------\n\n"
	return title
}

func getHelpText(s ScreenState) string {
	t := "\n\n\n"
	start := "[S]tart"
	pause := "[P]ause"
	reset := "[R]eset"
	back := "[B]ack"
	quit := "[Q]uit"
	enter := "[Enter] to start"
	goToEnter := "[E]nter expected"

	switch s {
	case Initial:
		t += fmt.Sprintf("%s\n%s", start, goToEnter)
	case Entering:
		t += fmt.Sprintf("%s\n%s", back, enter)
	case Running:
		t += fmt.Sprintf("%s\n%s", pause, reset)
	case Paused:
		t += fmt.Sprintf("%s\n%s", start, reset)
	case Stopped:
		t += fmt.Sprintf("%s\n%s", start, goToEnter)
	}
	t += fmt.Sprintf("\n%s", quit)
	return t
}

func renderRunningState(m model) string {
	s := getTitle(m.pomodoroTimerType)
	s += m.timer.View()
	if m.expected != 0 {
		a := "\n\nActual: %d\nExpected:%d"
		s += fmt.Sprintf(a, m.actual, m.expected)
	}
	s += getHelpText(m.screenState)
	return s
}

func renderPausedState(m model) string {
	s := getTitle(m.pomodoroTimerType)
	s += m.timer.View()
	s += getHelpText(m.screenState)
	return fmt.Sprintf(s, m.expected, m.actual)
}

func renderStoppedState(m model) string {
	s := getTitle(m.pomodoroTimerType)
	s += m.timer.View()
	if m.expected != 0 {
		a := "\n\nActual: %d\nExpected:%d"
		s += fmt.Sprintf(a, m.actual, m.expected)
	}
	s += getHelpText(m.screenState)
	return s
}

func renderEnteringState(m model) string {
	m.textFieldInput.Placeholder = "4"

	s := getTitle(m.pomodoroTimerType)
	s += "Expected 🍅s "
	s += m.textFieldInput.View()
	s += getHelpText(m.screenState)
	return s
}

func renderInitialState(m model) string {
	s := getTitle(m.pomodoroTimerType)
	s += "25:00"
	s += getHelpText(m.screenState)
	return s
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

func main() {
	notify = notificator.New(notificator.Options{
		DefaultIcon: "icon/default.png",
		AppName:     "ExPom",
	})

	f, err := tea.LogToFile("debug.log", "debug")
	if err != nil {
		log.Fatalln(`err:%w`, err)
	}
	defer f.Close()

	p := tea.NewProgram(initModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
