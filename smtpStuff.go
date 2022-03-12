package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jordan-wright/email"
	"html/template"
	"net/smtp"
	"os"
	"strconv"
	"strings"
	"time"
)

func SendSMTPEmail(too []string, subject string, html []byte, attachments ...string) (buf *bytes.Buffer, err error) {
	buf = new(bytes.Buffer)
	var to []string
	var failed []string
	for _, addr := range too {
		if CheckEmailCooldownUnsubscribe(addr) {
			to = append(to, addr)
		} else {
			failed = append(failed, addr)
		}
	}
	if len(to) == 0 {
		errorbuf := new(bytes.Buffer)
		fmt.Fprintf(errorbuf, "email(s) %s are on cooldown. No emails have been sent. Please try again in a little bit.", strings.Join(failed, ", "))
		return nil, errors.New(errorbuf.String())
	}
	fmt.Fprintln(buf, "SMTP: Sending email to", to, "subject", subject)
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
	err = em.SendWithStartTLS(FullConfig.SmtpDestination+":"+strconv.Itoa(FullConfig.SmtpPort), smtp.PlainAuth("", FullConfig.SmtpUsername, FullConfig.SmtpPassword, FullConfig.SmtpDestination), &tls.Config{ServerName: FullConfig.SmtpDestination})
	if err != nil {
		fmt.Fprintln(buf, "smtp error:", err.Error())
	} else {
		fmt.Fprintln(buf, "email.SendWithStartTLS completed.")
	}
	return
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

func SendTestSMTPEmail(to string) *bytes.Buffer {
	/*
		Used for debugging (./wvlist sendemail)
	*/
	text, _ := SendSMTPEmail([]string{to}, "Test email from "+FullConfig.Name, []byte("Hello. This is a test email, sent via ./wvlist smtp. If this email is recieved, that means your SMTP settings are entered correctly."))
	return text
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

type Apiv1SentSmtpEmailForEditUglyStr struct {
	Config ConfigStr
	Name   string
	Email  string
	Href   string
}

func Apiv1SentSmtpEmailForEditUgly(name, email, id, password string, info V1UploadEditUglyBodyOutputForEmail, diff []byte) error {
	/*
		Write the info json file and diff to temp files and send them with the email
	*/
	tmpFileInfo, err := os.CreateTemp("", "*.json")
	if err != nil {
		return err
	}
	cont, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return err
	}
	if _, err = tmpFileInfo.Write(cont); err != nil {
		return err
	}
	tmpFileInfo.Close()
	defer os.Remove(tmpFileInfo.Name())

	tmpFileDiff, err := os.CreateTemp("", "*.patch")
	if err != nil {
		return err
	}

	if _, err = tmpFileDiff.Write(diff); err != nil {
		return err
	}
	tmpFileDiff.Close()
	defer os.Remove(tmpFileDiff.Name())

	fmt.Println("Sending email to", name, "at", email)
	var san Apiv1SentSmtpEmailForEditUglyStr
	san.Config = *FullConfig
	san.Name = name
	san.Email = email
	san.Href = FullConfig.Hostname + "/api/v1/verifyedit/" + id + "/" + password + "/"

	/*
		Execute template
	*/

	buf := new(bytes.Buffer)

	tmp, err := template.ParseFiles("./template/apiv1editemail.html")
	if err != nil {
		return err
	}

	err = tmp.Execute(buf, san)
	if err != nil {
		return err
	}

	_, err = SendSMTPEmail([]string{email}, "Edit submission", buf.Bytes(), tmpFileInfo.Name(), tmpFileDiff.Name())

	return err
}

type Apiv1SendSmtpEmailForSubmitUglyStr struct {
	Config ConfigStr
	Name   string
	Email  string
	Href   string
}

func Apiv1SendSmtpEmailForSubmitUgly(san V1UploadUglySanitizedInput, fileIndex string, password string) error {
	/*
		Marshal contents of san
		to temp file and send that.
	*/

	tmpFile, err := CreateTemp("", "*.json")
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name())

	marshaled, err := json.MarshalIndent(san, "", "  ")
	if err != nil {
		return err
	}
	if _, err = tmpFile.Write(marshaled); err != nil {
		return err
	}
	tmpFile.Close()

	name := san.SubmitName
	emailAddress := san.SubmitEmail
	fmt.Println("sending email to", emailAddress)
	fmt.Println("name", name)

	htmlTemplate, err := template.ParseFiles("template/apiv1submissionemail.html")
	if err != nil {
		fmt.Println("htmlTemplate for email error", err.Error())
		return errors.New("Internal SMTP server error.")
	}
	var a Apiv1SendSmtpEmailForSubmitUglyStr
	a.Config = *FullConfig
	a.Name = name
	a.Email = emailAddress
	a.Href = FullConfig.Hostname + "/api/v1/verify/" + fileIndex + "/" + password + "/"
	var buf bytes.Buffer
	htmlTemplate.Execute(&buf, a)

	_, err = SendSMTPEmail([]string{emailAddress}, "Submission", buf.Bytes(), tmpFile.Name())
	return err

}
