/*
 *    Copyright (C) 2015 Stefan Luecke
 *
 *    This program is free software: you can redistribute it and/or modify
 *    it under the terms of the GNU Affero General Public License as published
 *    by the Free Software Foundation, either version 3 of the License, or
 *    (at your option) any later version.
 *
 *    This program is distributed in the hope that it will be useful,
 *    but WITHOUT ANY WARRANTY; without even the implied warranty of
 *    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *    GNU Affero General Public License for more details.
 *
 *    You should have received a copy of the GNU Affero General Public License
 *    along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 *    Authors: Stefan Luecke <glaxx@glaxx.net>
 */

package notification

import (
	"crypto/tls"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/fatih/structs"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"net/smtp"
	"strconv"
	"sync"
	"time"
)

type Mailconfig struct {
	Enabled       bool
	StartTLS      bool
	ServerAddress string
	Port          uint16
	Username      string
	Password      string
	EMailAddress  string
	Admin         string
	MaxAttempts   uint
}

func (m *Mailconfig) Verify() error {
	return nil
}

type MailNotificationService struct {
	status         chan int // status channel, 1 triggers an exit
	msg            chan mail
	deferred       *mgo.Collection
	mc             *Mailconfig
	wg             sync.WaitGroup
	nextDefAttempt time.Time
}

type mail struct {
	Id          bson.ObjectId `bson:"_id,omitempty"`
	header      header
	status      uint
	rcpt        string
	body        string
	nextAttempt time.Time
}

type header struct {
	From        string
	Date        time.Time
	Subject     string
	To          string
	ContentType string `mailheader:"Content-Type"`
	ReturnPath  string `mailheader:"Return-Path"`
}

func (h *header) toByte() []byte {
	var res string
	for _, f := range structs.Fields(h) {
		switch f.Value().(type) {
		case string:
			if t := f.Tag("mailheader"); t != "" {
				res = res + fmt.Sprintf("%v: %v \n", t, f.Value())
			} else {
				res = res + fmt.Sprintf("%v: %v \n", f.Name(), f.Value())
			}
			break
		case time.Time:
			res = res + fmt.Sprintf("%v: %v\n", f.Name(), f.Value().(time.Time).Format(time.RFC1123Z))
			break
		}
	}
	res = res + "\n"
	return []byte(res)
}

const (
	mailStatusNew = iota
	mailStatusPermanentFailure
	mailStatusAttemptOffset
)

func NewMailNotificationService(deferred *mgo.Collection, mailcfg *Mailconfig) *MailNotificationService {
	res := new(MailNotificationService)
	res.deferred = deferred
	res.mc = mailcfg
	res.status = make(chan int)
	res.msg = make(chan mail)
	res.wg.Add(1)
	go res.processQueue()
	return res
}

func (m *MailNotificationService) AddMailToQueue(rcpt, subject, text string) {
	ml := new(mail)
	ml.status = mailStatusNew
	ml.rcpt = rcpt
	ml.body = text
	ml.header.ContentType = "text/plain; charset=UTF-8"
	ml.header.Date = time.Now()
	ml.header.From = "lsmsd Notification Service <" + m.mc.EMailAddress + ">"
	ml.header.ReturnPath = m.mc.Admin
	ml.header.Subject = subject
	ml.header.To = rcpt
	m.msg <- *ml
}

func (m *MailNotificationService) Quit() {
	m.status <- 1
	m.wg.Wait()
}

func (m *MailNotificationService) processQueue() {
	defer m.wg.Done()
	hit := false
	for {
		select {
		case _ = <-m.status:
			return
		default:
			select {
			case ma := <-m.msg:
				hit = true
				err := m.sendMail(ma)
				if err != nil {
					//TODO: check for permanent failure
					m.deferSend(ma)
					log.Warn(err)
				}

			default:
			}
		}
		if time.Now().After(m.nextDefAttempt) {
			m.nextDefAttempt = time.Now().Add(6 * time.Hour)
			m.processDeferred()
		}

		if !hit {
			time.Sleep(200 * time.Millisecond)
		} else {
			hit = false
		}
	}
}

func (m *MailNotificationService) deferSend(ma mail) {
	err := m.deferred.Insert(ma)
	if err != nil {
		m.notifyAdmin(ma, err)
	}
}

func (m *MailNotificationService) processDeferred() {
	ma := make([]mail, 0)
	err := m.deferred.Find(nil).All(&ma)
	for i := 0; i != len(ma); i++ {
		err = m.sendMail(ma[i])
		if err != nil {
			m.notifyAdmin(ma[i], err)
		}
		err = m.deferred.Remove(bson.M{"_id": ma[i].Id})
		if err != nil {
			m.notifyAdmin(ma[i], err)
		}
	}
}

func (m *MailNotificationService) notifyAdmin(ma mail, err error) {
	ma.body = "Error while transmitting email to: " + ma.header.To + "\n" + err.Error() + "\n" + ma.body
	ma.header.Subject = "[ERROR]" + ma.header.Subject
	ma.header.To = m.mc.Admin
	ma.rcpt = m.mc.Admin
	er := m.sendMail(ma)
	if er != nil {
		log.Warn("Failed to notify admin: " + er.Error())
	}
}

func (m *MailNotificationService) sendMail(ma mail) error {
	auth := smtp.PlainAuth("", m.mc.Username, m.mc.Password, m.mc.ServerAddress)
	c, err := smtp.Dial(m.mc.ServerAddress + ":" + strconv.FormatUint(uint64(m.mc.Port), 10))
	if err != nil {
		return err
	}
	defer c.Close()

	if m.mc.StartTLS {
		if ok, _ := c.Extension("STARTTLS"); ok {
			conf := new(tls.Config)
			conf.ServerName = m.mc.ServerAddress
			err = c.StartTLS(conf)
			if err != nil {
				return err
			}
		} else {
			return errors.New("Server does not support StartTLS which is mandatory according to your settings")
		}
	}
	err = c.Auth(auth)
	if err != nil {
		return err
	}
	err = c.Mail(m.mc.EMailAddress)
	if err != nil {
		return err
	}
	err = c.Rcpt(ma.rcpt)
	if err != nil {
		return err
	}
	data, err := c.Data()
	if err != nil {
		return err
	}
	_, err = data.Write(ma.header.toByte())
	if err != nil {
		return err
	}
	_, err = data.Write([]byte(ma.body))
	if err != nil {
		return err
	}
	return data.Close()
}
