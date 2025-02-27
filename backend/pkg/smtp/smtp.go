package smtp

import (
	"net/smtp"
	"strconv"
)

type SMTPClient struct {
	Host     string
	Port     int
	Username string
	Password string
	FromAddr string
	FromName string
}

func NewSMTPClient(port int, host, uesrname, password, fromAddr, fromName string) (client *SMTPClient) {
	return &SMTPClient{Host: host, Port: port, Username: uesrname, Password: password, FromAddr: fromAddr, FromName: fromName}
}

func (sc *SMTPClient) SendMail(subject string, body string, toAddr string) (status bool, err error) {
	// create auth
	auth := smtp.PlainAuth("", sc.Username, sc.Password, sc.Host)
	// Convert port to string
	portStr := strconv.Itoa(sc.Port)

	msg := []byte(
		"Subject: " + subject + "\r\n" +
			"From: " + sc.FromName + " <" + sc.FromAddr + ">\r\n" +
			"To: " + toAddr + "\r\n" +
			"MIME-Version: 1.0\r\n" +
			"Content-Type: text/html; charset=\"utf-8\"\r\n" +
			"\r\n" +
			body,
	)

	err = smtp.SendMail(sc.Host+":"+portStr, auth, sc.FromAddr, []string{toAddr}, msg)
	if err != nil {
		return false, err
	}
	return true, nil
}
