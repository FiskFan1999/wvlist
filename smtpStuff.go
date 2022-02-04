package main

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"github.com/jordan-wright/email"
	"html/template"
	"net/smtp"
	"strconv"
	"strings"
	"time"
)

func SendSMTPEmail(too []string, subject string, html []byte, attachments ...string) {
	var to []string
	for _, addr := range too {
		if CheckEmailCooldownUnsubscribe(addr) {
			to = append(to, addr)
		}
	}
	if len(to) == 0 {
		return
	}
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

var EmailCoolDown map[string]time.Time

func CheckEmailCooldownUnsubscribe(a string) bool {
	/*
		Sanitize input email, to protect against spamming
		using periods in the to field (gmail ignores periods)

		Note that this will also remove the period from the
		domain name (@gmail.com -> @gmailcom) but this is
		ok because the check will function the same.
	*/
	addr := strings.ReplaceAll(a, ".", "")
	if !CheckEmailUnsibscribe(addr) {
		return false
	}

	previousCall, ok := EmailCoolDown[addr]
	if !ok || // item doesn't exist.
		time.Since(previousCall).Minutes() > float64(1) { //
		// Email is valid
		EmailCoolDown[addr] = time.Now()
		return true
	} else {
		fmt.Println("smtp: can't send email to", a, "due to cooldown")
		return false
	}

	//return true
}

func CheckEmailUnsibscribe(addr string) bool {
	return true
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

type Apiv1SendSmtpEmailForSubmitUglyStr struct {
	Config ConfigStr
	Name   string
	Href   string
}

func Apiv1SendSmtpEmailForSubmitUgly(san V1UploadUglySanitizedInput, fileIndex string, password string) {
	/*
		TO DO
	*/
	name := san.SubmitName
	emailAddress := san.SubmitEmail
	fmt.Println("sending email to", emailAddress)
	fmt.Println("name", name)

	htmlTemplate, err := template.ParseFiles("template/apiv1submissionemail.html")
	if err != nil {
		fmt.Println("htmlTemplate for email error", err.Error())
		return
	}
	var a Apiv1SendSmtpEmailForSubmitUglyStr
	a.Config = *FullConfig
	a.Name = name
	a.Href = "http://127.0.0.1:6060/api/v1/verify/" + fileIndex + "/" + password + "/"
	var buf bytes.Buffer
	htmlTemplate.Execute(&buf, a)

	SendSMTPEmail([]string{emailAddress}, "Submission", buf.Bytes())

}
