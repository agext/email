package email

import (
	"errors"
	"net/smtp"
	"strconv"
	"sync"
)

// Sender represents the SMTP credentials along with the (optional) Address of a sender.
type Sender struct {
	host     string
	port     int
	username string
	password string
	address  *Address
}

var (
	defaultSender      *Sender
	defaultSenderMutex sync.RWMutex
)

// NewSender creates a new Sender from the provided information.
//
// The `host` may include a port number, which defaults to 25. That is, "example.com"
// and "example.com:25" are equivalent.
// The `addr` parameters are optional and may be either an email address or a name followed by an
// email address.
func NewSender(host, user, pass string, addr ...string) (*Sender, error) {
	port := 0
	for i, l := 0, len(host); i < l; i++ {
		if host[i] == ':' {
			for _, digit := range host[i+1:] {
				if digit < '0' || digit > '9' {
					return nil, errors.New("NewSender: invalid port number: " + host)
				}
				port = port*10 + int(digit-'0')
			}
			host = host[:i]
			break
		}
	}
	if port == 0 {
		port = 25
	}
	if user == "" {
		return nil, errors.New("NewSender: empty username: " + user)
	}
	if pass == "" {
		return nil, errors.New("NewSender: empty password: " + pass)
	}
	var (
		address *Address
		err     error
	)
	switch len(addr) {
	case 2:
		address, err = NewAddress(addr[0], addr[1])
	case 1:
		address, err = NewAddress("", addr[0])
	}
	if err != nil {
		return nil, errors.New("NewSender: " + err.Error())
	}
	return &Sender{host, port, user, pass, address}, nil
}

// SetDefault sets the receiver as the default sender.
func (s *Sender) SetDefault() *Sender {
	defaultSenderMutex.Lock()
	defaultSender = s
	defaultSenderMutex.Unlock()
	return s
}

// Send composes the provided message using the `data`, and sends it.
func (s *Sender) Send(msg *Message, data interface{}) error {
	if msg == nil {
		return errors.New("Sender.Send: no message to send")
	}
	body := msg.setSender(s).Compose(data)
	if msg.HasErrors() {
		return errors.New("Sender.Send: failed to compose message")
	}
	go smtp.SendMail(
		s.host+":"+strconv.Itoa(s.port),
		smtp.PlainAuth(
			"",
			s.username,
			s.password,
			s.host,
		),
		msg.FromAddr(),
		msg.RecipientAddrs(),
		body,
	)
	return nil
}

// Send composes the provided message using the `data`, and sends it using the default Sender.
func Send(msg *Message, data interface{}) error {
	defaultSenderMutex.RLock()
	defer defaultSenderMutex.RUnlock()
	sender := defaultSender
	if sender == nil {
		return errors.New("Send: no default sender")
	}
	return sender.Send(msg, data)
}
