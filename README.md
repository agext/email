# A Go package for composing and sending email messages

[![Release](https://img.shields.io/github/release/agext/email.svg?style=flat&colorB=eebb00)](https://github.com/agext/email/releases/latest)
[![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat)](https://godoc.org/github.com/agext/email) 
[![Build Status](https://travis-ci.org/agext/email.svg?branch=master&style=flat)](https://travis-ci.org/agext/email)
[![Coverage Status](https://coveralls.io/repos/github/agext/email/badge.svg?style=flat)](https://coveralls.io/github/agext/email)
[![Go Report Card](https://goreportcard.com/badge/github.com/agext/email?style=flat)](https://goreportcard.com/report/github.com/agext/email)


This package implements a simple yet powerful API for email message composition and sending via an authenticating SMTP server, in [Go](http://golang.org).

## Project Status

v0.2.1 Edge: Breaking changes to the API unlikely but possible until the v1.0 release. May be robust enough to use in production, though provided on "AS IS" basis. Vendoring recommended.

This package is under active development. If you encounter any problems or have any suggestions for improvement, please [open an issue](https://github.com/agext/email/issues). Pull requests are welcome.

## Overview

This package provides a fluid API for composing and sending email messages.

Any application needs one (and usually only one) `Sender`, representing an SMTP account connection information plus a sender `Address`.

A `Message` contains all the information required to create an email message. Usually, at least some parts of each message are templates to be filled with data from the rest of the application. Also, it is often convenient to define base messages on program initialization, and clone them for fine-tuning and send-out when needed.

Below is an overly-simplified example, to showcase the basic functionality. Note that by not defining From: and To: addresses for a message, they default to the sender address, which is convenient for system messaging.

```go
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

	// send the message after composing it with the provided data
	err := sender.Send(msg, data)

	if err != nil {
		log.Println(err, msg.Errors())
	}

	return err
}

```

## Installation

```
go get github.com/agext/email
```

## License

Package email is released under the Apache 2.0 license. See the [LICENSE](LICENSE) file for details.
