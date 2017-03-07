/*
Package email provides a fluid API for composing and sending email messages.

Any application needs one (and usually only one) Sender, representing an SMTP account connection information plus a sender Address.

A Message contains all the information required to create an email message. Usually, at least some parts of each message are templates to be filled with data from the rest of the application. Also, it is often convenient to define base messages on program initialization, and clone them for fine-tuning and send-out when needed.

Below is an overly-simplified example, to showcase the basic functionality. Note that by not defining From: and To: addresses for a message, they default to the sender address, which is convenient for system messaging.

	package main

	import (
		"log"

		"github.com/agext/email"
	)

	var (
		host               = "smtp.example.com"
		user               = "username"
		pass               = "password"
		name               = "Application mail"
		addr               = "app@example.com"
		sender             *email.Sender
		contactFormMessage *email.Message
	)

	func main() {
		var err error
		// create a sender with a given configuration
		sender, err = email.NewSender(host, user, pass, name, addr)
		if err != nil {
			log.Fatalln("invalid sender configuration: " + err.Error())
		}

		// create a message from scratch, to be used as a base for the actual messages
		contactFormMessage = email.NewMessage(nil).
			SubjectTemplate("Contact form message from {{.first}} {{.last}}").
			TextTemplate(`
	First Name:   {{.first}}
	Last Name:    {{.last}}
	Phone:        {{.phone}}
	Email:        {{.email}}
	`)

		// ...
	}

	func sendContact(data map[string]interface{}) error {
		// create a message from the base we created in main() - basically, clone all its data
		msg := email.NewMessage(contactFormMessage)

		// customize / adapt the message as needed...
		// if the base messages are well thought out, the need should be minimal, if at all
		// as an example, we could set the To: address based on the form data, to dispatch the messages
		// to the person in the company who is most capable to handle it.

		// send the message after composing it with the provided date
		err := sender.Send(msg, data)

		if err != nil {
			log.Println(err, msg.Errors())
		}

		return err
	}

*/
package email
