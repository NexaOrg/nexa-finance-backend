package utils

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"time"
)

type MailServer struct {
	Server string
	Port   string
	Auth   smtp.Auth
	From   string
}

func NewMailServer(server, port, from, user, password string) *MailServer {
	auth := smtp.PlainAuth("", user, password, server)
	return &MailServer{
		Server: server,
		Port:   port,
		Auth:   auth,
		From:   from,
	}
}

func (ms *MailServer) SendEmailHTML(subject, html string, to []string) error {
	if len(to) == 0 {
		return fmt.Errorf("no recipients provided")
	}

	client, err := ms.connect()
	if err != nil {
		return fmt.Errorf("failed to connect to mail server: %s", err)
	}
	defer client.Quit()

	if err = client.Mail(ms.From); err != nil {
		return fmt.Errorf("error setting sender: %s", err)
	}

	for _, recipient := range to {
		if err = client.Rcpt(recipient); err != nil {
			return fmt.Errorf("error setting recipient %s: %s", recipient, err)
		}
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("error initiating data: %s", err)
	}

	headers := fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-version: 1.0\r\nContent-Type: text/html; charset=\"UTF-8\"\r\n\r\n",
		ms.From,
		to[0],
		subject,
	)

	msg := headers + html

	_, err = w.Write([]byte(msg))
	if err != nil {
		return fmt.Errorf("error writing email data: %s", err)
	}

	err = w.Close()
	if err != nil {
		return fmt.Errorf("error closing writer: %s", err)
	}

	return nil
}

func (ms *MailServer) connect() (*smtp.Client, error) {
	conn, err := ms.newConnection()
	if err != nil {
		return nil, err
	}

	client, err := ms.newClient(conn)
	if err != nil {
		conn.Close()
		return nil, err
	}

	return client, nil
}

func (ms *MailServer) newConnection() (*tls.Conn, error) {
	addr := ms.Server + ":" + ms.Port
	dialer := &net.Dialer{
		Timeout: 15 * time.Second,
	}

	tlsConfig := &tls.Config{
		MinVersion:         tls.VersionTLS12,
		InsecureSkipVerify: false,
	}

	conn, err := tls.DialWithDialer(dialer, "tcp", addr, tlsConfig)
	if err != nil {
		return nil, fmt.Errorf("TLS connection error: %s", err)
	}

	return conn, nil
}

func (ms *MailServer) newClient(conn *tls.Conn) (*smtp.Client, error) {
	client, err := smtp.NewClient(conn, ms.Server)
	if err != nil {
		return nil, fmt.Errorf("error creating SMTP client: %s", err)
	}

	if err = client.Auth(ms.Auth); err != nil {
		return nil, fmt.Errorf("SMTP authentication error: %s", err)
	}

	return client, nil
}
