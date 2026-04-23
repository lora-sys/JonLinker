package service

import (
	"bytes"
	"fmt"
	"joblinker/internal/agent"
	"joblinker/internal/model"
	"net/smtp"
	"os"
	"time"
)

type EmailService struct {
	smtpHost string
	smtpPort string
	from     string
	password string
}

func NewEmailService() *EmailService {
	return &EmailService{
		smtpHost: os.Getenv("SMTP_HOST"),
		smtpPort: os.Getenv("SMTP_PORT"),
		from:     os.Getenv("SMTP_FROM"),
		password: os.Getenv("SMTP_PASSWORD"),
	}
}

func (s *EmailService) SendInterviewInvitation(interview *model.Interview, recipientEmail string) error {
	if s.smtpHost == "" {
		// Skip sending in development if SMTP not configured
		return nil
	}

	icsContent, err := s.generateICS(interview)
	if err != nil {
		return err
	}

	subject := fmt.Sprintf("Interview Scheduled - %s", interview.ScheduledAt.Format("Jan 2, 2006 at 3:04 PM"))
	body := s.buildEmailBody(interview)

	msg := s.buildMessage(recipientEmail, subject, body, icsContent)

	addr := fmt.Sprintf("%s:%s", s.smtpHost, s.smtpPort)
	auth := smtp.PlainAuth("", s.from, s.password, s.smtpHost)

	return smtp.SendMail(addr, auth, s.from, []string{recipientEmail}, msg)
}

func (s *EmailService) SendReminder(interview *model.Interview, recipientEmail string) error {
	if s.smtpHost == "" {
		return nil
	}

	subject := fmt.Sprintf("Reminder: Interview Tomorrow at %s", interview.ScheduledAt.Format("3:04 PM"))
	body := s.buildReminderBody(interview)

	msg := s.buildMessage(recipientEmail, subject, body, "")

	addr := fmt.Sprintf("%s:%s", s.smtpHost, s.smtpPort)
	auth := smtp.PlainAuth("", s.from, s.password, s.smtpHost)

	return smtp.SendMail(addr, auth, s.from, []string{recipientEmail}, msg)
}

func (s *EmailService) generateICS(interview *model.Interview) (string, error) {
	endTime := interview.ScheduledAt.Add(time.Hour)
	return agent.GenerateICS(
		"Job Interview",
		"Interview scheduled via JobLinker",
		interview.ScheduledAt,
		endTime,
		interview.Location,
	), nil
}

func (s *EmailService) buildEmailBody(interview *model.Interview) string {
	return fmt.Sprintf(`Dear Candidate,

Your interview has been scheduled for:

Date: %s
Time: %s
Format: %s
Location: %s

Please find the calendar invitation attached.

Best regards,
JobLinker Team`,
		interview.ScheduledAt.Format("Monday, January 2, 2006"),
		interview.ScheduledAt.Format("3:04 PM MST"),
		interview.Format,
		interview.Location,
	)
}

func (s *EmailService) buildReminderBody(interview *model.Interview) string {
	return fmt.Sprintf(`Dear Candidate,

This is a friendly reminder that your interview is scheduled for tomorrow:

Date: %s
Time: %s
Format: %s
Location: %s

Please ensure you are prepared and log in a few minutes early.

Best regards,
JobLinker Team`,
		interview.ScheduledAt.Format("Monday, January 2, 2006"),
		interview.ScheduledAt.Format("3:04 PM MST"),
		interview.Format,
		interview.Location,
	)
}

func (s *EmailService) buildMessage(to, subject, body, icsContent string) []byte {
	var msg bytes.Buffer

	msg.WriteString("From: " + s.from + "\r\n")
	msg.WriteString("To: " + to + "\r\n")
	msg.WriteString("Subject: " + subject + "\r\n")
	msg.WriteString("MIME-Version: 1.0\r\n")

	if icsContent != "" {
		msg.WriteString("Content-Type: multipart/mixed; boundary=boundary\r\n\r\n")
		msg.WriteString("--boundary\r\n")
		msg.WriteString("Content-Type: text/plain; charset=utf-8\r\n\r\n")
		msg.WriteString(body + "\r\n\r\n")
		msg.WriteString("--boundary\r\n")
		msg.WriteString("Content-Type: text/calendar; method=REQUEST\r\n")
		msg.WriteString("Content-Transfer-Encoding: 7bit\r\n\r\n")
		msg.WriteString(icsContent + "\r\n")
		msg.WriteString("--boundary--\r\n")
	} else {
		msg.WriteString("Content-Type: text/plain; charset=utf-8\r\n\r\n")
		msg.WriteString(body)
	}

	return msg.Bytes()
}
