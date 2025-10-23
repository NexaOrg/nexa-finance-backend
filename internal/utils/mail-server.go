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
	client, err := ms.connect()
	if err != nil {
		return fmt.Errorf("failed to connect to mail server: %s", err)
	}

	for _, recipient := range to {
		if err = client.Rcpt(recipient); err != nil {
			return fmt.Errorf("error setting recipient: %s", err)
		}
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("error sending data: %s", err)
	}

	headers := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";"
	msg := "Subject: " + subject + "\n" + headers + "\n\n" + html

	_, err = w.Write([]byte(msg))
	if err != nil {
		return fmt.Errorf("error writing email data: %s", err)
	}

	err = w.Close()
	if err != nil {
		return fmt.Errorf("error closing writer: %s", err)
	}

	err = client.Quit()
	if err != nil {
		return fmt.Errorf("error quit client: %s", err)
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
		return nil, err
	}

	return client, nil
}

func (ms *MailServer) newConnection() (*tls.Conn, error) {
	addr := ms.Server + ":" + ms.Port

	dialer := &net.Dialer{
		Timeout: 10 * time.Second,
	}

	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS13,
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

	if err = client.Mail(ms.From); err != nil {
		return nil, fmt.Errorf("error setting sender: %s", err)
	}

	return client, nil
}
