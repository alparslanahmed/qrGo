package email

import (
	"alparslanahmed/qrGo/config"
	"io/ioutil"
	"log"
	"net/smtp"
	"os"
	"strings"
)

func SendHTMLEmail(to string, subject string, body string, file string) error {
	// SendHTMLEmail sends an email with HTML content
	from := config.Config("MAIL_SENDER")
	pass := config.Config("MAIL_PASSWORD")

	// Read the HTML file content
	htmlContent, err := LocalHTML(file)

	htmlContent = strings.Replace(htmlContent, "{{body}}", body, -1)

	if err != nil {
		log.Printf("error reading HTML file: %s", err)
		return err
	}

	msg := "From: Enfes Menü" + " <" + from + ">\n" +
		"To: " + to + "\n" +
		"Subject: " + subject + "\n" +
		"MIME-version: 1.0;\n" +
		"Content-Type: text/html; charset=\"UTF-8\";\n\n" +
		string(htmlContent)

	err = smtp.SendMail("smtp.gmail.com:587",
		smtp.PlainAuth("", from, pass, "smtp.gmail.com"),
		from, []string{to}, []byte(msg))

	if err != nil {
		log.Printf("smtp error: %s", err)
		return err
	}

	return nil
}

func LocalHTML(filepath string) (string, error) {
	// Get the project folder
	projectFolder, err := getProjectFolder()
	if err != nil {
		return "", err
	}
	// Read the HTML file content
	content, err := ioutil.ReadFile(projectFolder + "/email" + filepath)
	if err != nil {
		return "", err
	}

	return string(content), nil
}

func getProjectFolder() (string, error) {
	// Get the current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// Traverse up the directory tree to find the project folder
	projectFolder := cwd

	return projectFolder, nil
}
