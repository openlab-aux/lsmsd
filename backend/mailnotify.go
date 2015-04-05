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

package backend

import (
	//	log "github.com/Sirupsen/logrus"
	"gopkg.in/mgo.v2"
	//	"gopkg.in/mgo.v2/bson"
	//	"net/smtp"
	"sync"
	//	"time"
)

type Mailconfig struct {
	Enabled       bool
	StartTLS      bool
	ServerAddress string
	Username      string
	Password      string
	EMailAddress  string
	Admins        []string
}

func (m *Mailconfig) Verify() error {
	return nil
}

type MailNotificationService struct {
	status   chan int // status channel, 1 triggers an exit
	msg      chan mail
	user     *mgo.Collection
	deferred *mgo.Collection
	mc       *Mailconfig
	wg       sync.WaitGroup
}

type mail struct {
	rcpt string
	text string
}

func NewMailNotificationService(user, deferred *mgo.Collection, mailcfg *Mailconfig) *MailNotificationService {
	res := new(MailNotificationService)
	res.user = user
	res.deferred = deferred
	res.mc = mailcfg
	return res
}

func (m *MailNotificationService) AddMailToQueue(rcpt, text string) {
}

func (m *MailNotificationService) processQueue() {
}

func (m *MailNotificationService) deferSend() {
}

func (m *MailNotificationService) processDeferred() {
}

func (m *MailNotificationService) notifyAdmin() {
}

func (m *MailNotificationService) sendMail() {
}
