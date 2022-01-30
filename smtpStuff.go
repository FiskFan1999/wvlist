package main

import (
	"crypto/tls"
	"fmt"
	"github.com/jordan-wright/email"
	"net/smtp"
	"strconv"
)

func SendSMTPEmail(to []string, subject string, html []byte, attachments ...string) {
	fmt.Println("SMTP: Sending email to", to, "subject", subject)
	em := email.NewEmail()
	em.From = "server@wvlist.net"
	em.To = to
	if len(FullConfig.SmtpAdminBCC) != 0 {
		em.Bcc = FullConfig.SmtpAdminBCC
	}
	em.Subject = subject
	em.HTML = html
	for _, att := range attachments {
		em.AttachFile(att)
	}
	err := em.SendWithStartTLS(FullConfig.SmtpDestination+":"+strconv.Itoa(FullConfig.SmtpPort), smtp.PlainAuth("", FullConfig.SmtpUsername, FullConfig.SmtpPassword, FullConfig.SmtpDestination), &tls.Config{ServerName: FullConfig.SmtpDestination})
	if err != nil {
		fmt.Println("smtp error:", err.Error())
	} else {
		fmt.Println("email.SendWithStartTLS completed.")
	}
}

func SendTestSMTPEmail(to string) {
	/*
		Used for debugging (./wvlist sendemail)
	*/
	SendSMTPEmail([]string{to}, "Test email from "+FullConfig.Name, []byte("Hello. This is a test email, sent via ./wvlist smtp. If this email is recieved, that means your SMTP settings are entered correctly."))
	/*
		em := email.NewEmail()
		em.From = "server@wvlist.net"
		em.To = []string{to}
		em.Subject = "Test email from " + FullConfig.Name
		em.Text = []byte("Hello. This is a test email, sent via ./wvlist smtp. If this email is recieved, that means your SMTP settings are entered correctly.")
		err := em.SendWithStartTLS(FullConfig.SmtpDestination+":"+strconv.Itoa(FullConfig.SmtpPort), smtp.PlainAuth("", FullConfig.SmtpUsername, FullConfig.SmtpPassword, FullConfig.SmtpDestination), &tls.Config{ServerName: FullConfig.SmtpDestination})
		if err != nil {
			fmt.Println("smtp error:", err.Error())
		} else {
			fmt.Println("email.SendWithStartTLS completed with no errors")
		}
	*/
}

func Apiv1SendSmtpEmailForSubmitUgly(san V1UploadUglySanitizedInput) {
	/*
		TO DO
	*/
	name := san.SubmitName
	emailAddress := san.SubmitEmail
	fmt.Println("sending email to", emailAddress)
	fmt.Println("name", name)
}
