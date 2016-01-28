package main

import (
	"strconv"
	"strings"
)

type Message struct {
	Unparsed   string
	Command    string
	FullSender string
	Sender     string
	Forum      string
	Args       []string
	Text       string
}

func NewMessage(v string) (Message, error) {
	var m Message
	var parts []string
	var lhs string

	m.Unparsed = v
	parts = strings.SplitN(v, " :", 2)
	if len(parts) == 2 {
		lhs = parts[0]
		m.Text = parts[1]
	} else {
		lhs = v
		m.Text = ""
	}

	m.FullSender = "."
	m.Forum = "."
	m.Sender = "."

	parts = strings.Split(lhs, " ")
	if parts[0][0] == ':' {
		m.FullSender = parts[0][1:]
		parts = parts[1:]

		n, u, _ := SplitTarget(m.FullSender)
		if u != "" {
			m.Sender = n
		}
	}

	m.Command = strings.ToUpper(parts[0])
	switch m.Command {
	case "PRIVMSG", "NOTICE":
		switch {
		case IsChannel(parts[1]):
			m.Forum = parts[1]
		case m.FullSender == ".":
			m.Forum = parts[1]
		default:
			m.Forum = m.Sender
		}
	case "PART", "MODE", "TOPIC", "KICK":
		m.Forum = parts[1]
		m.Args = parts[2:]
	case "JOIN":
		if len(parts) == 1 {
			m.Forum = m.Text
			m.Text = ""
		} else {
			m.Forum = parts[1]
		}
	case "INVITE":
		if m.Text != "" {
			m.Forum = m.Text
			m.Text = ""
		} else {
			m.Forum = parts[2]
		}
	case "NICK":
		if len(parts) > 1 {
			m.Sender = parts[1]
			m.Args = parts[2:]
		} else {
			m.Sender = m.Text
			m.Text = ""
			m.Args = parts[1:]
		}
		m.Forum = m.Sender
	case "353":
		m.Forum = parts[3]
	default:
		numeric, _ := strconv.Atoi(m.Command)
		if numeric >= 300 {
			if len(parts) > 2 {
				m.Forum = parts[2]
			}
		}
		m.Args = parts[1:]
	}

	return m, nil
}

func (m Message) String() string {
	return m.Unparsed
}

func SplitTarget(s string) (string, string, string) {
	var parts []string

	parts = strings.SplitN(s, "!", 2)
	if len(parts) == 1 {
		return s, "", ""
	}
	nick := parts[0]
	parts = strings.SplitN(parts[1], "@", 2)
	if len(parts) == 1 {
		return s, "", ""
	}
	return nick, parts[0], parts[1]
}

func IsChannel(s string) bool {
	if s == "" {
		return false
	}

	switch s[0] {
	case '#', '+', '!', '&':
		return true
	default:
		return false
	}
}
