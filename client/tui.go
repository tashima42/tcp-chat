package client

import (
	"fmt"
	"net"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type (
	errMsg error
	newMsg struct {
		username string
		value    string
	}
)

var (
	focusedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	blurredStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	noStyle      = lipgloss.NewStyle()
	helpStyle    = blurredStyle.Copy()
)

type model struct {
	viewport      viewport.Model
	messages      []string
	registered    bool
	usernameInput textinput.Model
	textarea      textarea.Model
	senderStyle   lipgloss.Style
	conn          *net.Conn
	err           error
}

func initialModel(conn *net.Conn) model {
	ta := textarea.New()
	ta.Placeholder = "Send a message..."
	ta.Focus()

	ta.Prompt = "┃ "
	ta.CharLimit = 280

	ta.SetWidth(30)
	ta.SetHeight(3)

	// Remove cursor line styling
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()

	ta.ShowLineNumbers = false

	vp := viewport.New(30, 5)
	vp.SetContent(`Welcome to the chat room!
Type a message and press Enter to send.`)

	ta.KeyMap.InsertNewline.SetEnabled(false)

	ti := textinput.New()
	ti.Placeholder = "Input your username"
	ti.Focus()
	ti.CharLimit = 20
	ti.Width = 20

	return model{
		textarea:      ta,
		messages:      []string{},
		registered:    false,
		usernameInput: ti,
		viewport:      vp,
		senderStyle:   lipgloss.NewStyle().Foreground(lipgloss.Color("5")),
		conn:          conn,
		err:           nil,
	}
}

func (m model) Init() tea.Cmd {
	return textarea.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.registered {
		return updateChat(m, msg)
	}
	return updateRegister(m, msg)
}

func updateRegister(m model, msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit
		case "enter":
			register(*m.conn, m.usernameInput.Value())
			m.registered = true
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.usernameInput, cmd = m.usernameInput.Update(msg)

	return m, cmd
}

func updateChat(m model, msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
	)

	m.textarea, tiCmd = m.textarea.Update(msg)
	m.viewport, vpCmd = m.viewport.Update(msg)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			fmt.Println(m.textarea.Value())
			return m, tea.Quit
		case tea.KeyEnter:
			m.messages = append(m.messages, m.senderStyle.Render("[you]: ")+m.textarea.Value())
			m.viewport.SetContent(strings.Join(m.messages, "\n"))
			sendMessage(*m.conn, m.textarea.Value())
			m.textarea.Reset()
			m.viewport.GotoBottom()
		}
	case newMsg:
		m.messages = append(m.messages, m.senderStyle.Render(fmt.Sprintf("[%s]: %s", msg.username, msg.value)))
		m.viewport.SetContent(strings.Join(m.messages, "\n"))
		m.viewport.GotoBottom()

	case errMsg:
		m.err = msg
		return m, nil
	}

	return m, tea.Batch(tiCmd, vpCmd)
}

func (m model) View() string {
	if m.registered {
		return chatView(m)
	}
	return registerView(m)
}

func chatView(m model) string {
	return fmt.Sprintf(
		"%s\n\n%s",
		m.viewport.View(),
		m.textarea.View(),
	) + "\n\n"
}

func registerView(m model) string {
	var b strings.Builder

	b.WriteString(m.usernameInput.View())
	b.WriteRune('\n')
	b.WriteRune('\n')
	b.WriteString(helpStyle.Render("press enter to submit"))
	b.WriteRune('\n')
	b.WriteString(helpStyle.Render("press esc o ctrl+c to exit"))

	return b.String()
}
