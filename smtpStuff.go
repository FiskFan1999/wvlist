package main

import (
	"fmt"
)

func Apiv1SendSmtpEmailForSubmitUgly(san V1UploadUglySanitizedInput) {
	/*
		TO DO
	*/
	name := san.SubmitName
	emailAddress := san.SubmitEmail
	fmt.Println("sending email to", emailAddress)
	fmt.Println("name", name)
}
