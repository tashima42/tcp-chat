package client

import (
	"fmt"
	"net"
	"slices"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/tashima42/tcp-chat/types"
)

type errMsg error

var (
	blurredStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	senderStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("5"))
	receiverStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("4"))
	titleStyle    = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Right = "├"
		return lipgloss.NewStyle().BorderStyle(b).Padding(0, 1).Margin(0)
	}()
	sideStyle = func() lipgloss.Style {
		b := lipgloss.NormalBorder()
		return lipgloss.NewStyle().BorderStyle(b).Padding(0, 1)
	}()
	helpStyle = blurredStyle.Copy()
)

type model struct {
	viewport      viewport.Model
	messages      []string
	users         map[string]types.User
	usersLength   int
	registered    bool
	usernameInput textinput.Model
	messageInput  textinput.Model
	viewportReady bool
	height        int
	conn          *net.Conn
	err           error
}

func initialModel(conn *net.Conn) model {
	mi := textinput.New()
	mi.Placeholder = "Send a message..."
	mi.Focus()

	mi.Prompt = "┃ "
	mi.CharLimit = 560

	vp := viewport.New(100, 5)
	vp.SetContent(`Welcome to the chat room!
Type a message and press Enter to send.`)

	ti := textinput.New()
	ti.Placeholder = "Input your username"
	ti.Focus()
	ti.CharLimit = 20
	ti.Width = 20

	return model{
		messageInput:  mi,
		messages:      []string{},
		users:         map[string]types.User{},
		usersLength:   0,
		registered:    false,
		usernameInput: ti,
		viewport:      vp,
		viewportReady: false,
		height:        10,
		conn:          conn,
		err:           nil,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(textarea.Blink, tea.EnterAltScreen)
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

	case tea.WindowSizeMsg:
		headerHeight := lipgloss.Height(m.headerView())
		footerHeight := lipgloss.Height(m.footerView())
		verticalMarginHeight := headerHeight + footerHeight
		m.height = msg.Height

		if !m.viewportReady {
			m.viewport = viewport.New(msg.Width, msg.Height-verticalMarginHeight)
			m.viewport.YPosition = headerHeight
			m.viewport.HighPerformanceRendering = false
			m.viewportReady = true
			m.viewport.YPosition = headerHeight + 1
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - verticalMarginHeight
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

	m.messageInput, tiCmd = m.messageInput.Update(msg)
	m.viewport, vpCmd = m.viewport.Update(msg)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			fmt.Println(m.messageInput.Value())
			return m, tea.Quit
		case tea.KeyEnter:
			m.messages = append(m.messages, senderStyle.Render("[you]: ")+m.messageInput.Value())
			m.viewport.SetContent(strings.Join(m.messages, "\n"))
			sendMessage(*m.conn, m.messageInput.Value())
			m.messageInput.Reset()
			m.viewport.GotoBottom()
		}
	case types.Users:
		m.users = map[string]types.User{}
		m.usersLength = len(msg)
		for _, u := range msg {
			m.users[u.ID] = u
		}
		m.viewport.SetContent(strings.Join(m.messages, "\n"))
		m.viewport.GotoBottom()
		return m, nil
	case types.Message:
		user := m.users[msg.UserID]
		m.messages = append(m.messages, receiverStyle.Render(fmt.Sprintf("[%s]: ", user.Username))+msg.Value)
		m.viewport.SetContent(strings.Join(m.messages, "\n"))
		m.viewport.GotoBottom()
		return m, nil
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
	if !m.viewportReady {
		return "\n  Initializing..."
	}
	chat := fmt.Sprintf("%s\n%s\n%s", m.headerView(), m.viewport.View(), m.footerView())
	return lipgloss.JoinHorizontal(lipgloss.Top, m.sideView(), chat)

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

func (m model) sideView() string {
	users := []string{}
	for _, v := range m.users {
		users = append(users, v.Username)
	}
	usersList := ""
	slices.Sort(users)
	usersList = strings.Join(users, "\n")
	return sideStyle.MaxHeight(m.height).Height(m.height - m.usersLength).Render(usersList)
}

func (m model) headerView() string {
	title := titleStyle.Render("TCP Chat")
	line := strings.Repeat("─", max(0, m.viewport.Width-lipgloss.Width(title)))
	return lipgloss.JoinHorizontal(lipgloss.Center, title, line)
}

func (m model) footerView() string {
	line := strings.Repeat("─", max(0, m.viewport.Width))
	input := m.messageInput.View()
	return lipgloss.JoinVertical(lipgloss.Left, line, input, line)
}
