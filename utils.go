package main

import gomail "gopkg.in/gomail.v2"
import "fmt"

func sendMail(s TargetStatus, c Config) {

	for _, to := range c.Mails {
		m := gomail.NewMessage()
		m.SetHeader("From", c.Sender)
		m.SetHeader("To", to)
		//m.SetAddressHeader("Cc", "dan@example.com", "Dan")
		m.SetHeader("Subject", "Pingo Status")
		if s.Online {
			message := fmt.Sprintf("%s is ok", s.Target.Name)
			m.SetBody("text/html", message)
		} else {
			message := fmt.Sprintf("%s is ok", s.Target.Name)
			m.SetBody("text/html", message)
		}
		//m.Attach("/home/Alex/lolcat.jpg")

		d := gomail.Dialer{Host: "mailpi.smals.be", Port: 25}
		if err := d.DialAndSend(m); err != nil {
			panic(err)
		}
	}
}
