package main

const (
	SubmitPageText = "Thank you for your submission! If you feel comfortable submitting incipits, please test your incipits in the lilypond sandbox to ensure that they will look correctly. (Incipits are not required, so do not feel pressured to include them unless you understand the syntax and know what you are doing.)"

	LilypondPageText = "Welcome to the Lilypond Sandbox! Please enter the incipit as if it were inside the \\score block of the lilypond input file (as demonstrated below). Use this page to ensure that your incipit looks as intended and does not take up too much space. As for the size, kindly keep your length of incipit to around one measure, or approximately ten to twelve notes in a row. Please note that incipits are optional, so do not feel pressured to enter them unless you know what you are doing. Consider the following examples for some inspiration for the length of the incipit."
)

var LilyIncipitExamples []string = []string{
	"/lilypondexamples/1.png",
	"/lilypondexamples/2.png",
	"/lilypondexamples/3.png",
}
