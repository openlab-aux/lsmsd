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

package webservice_test

import (
	log "github.com/Sirupsen/logrus"
	"github.com/emicklei/go-restful"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	db "github.com/openlab-aux/lsmsd/database"
	"github.com/openlab-aux/lsmsd/webservice"
	"gopkg.in/mgo.v2"
	"strconv"
	"strings"
	"testing"
)

func TestWebservice(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Webservice Suite")
}

var _ = BeforeSuite(func() {
	log.SetLevel(log.FatalLevel)
})

func newTestContainer() (*mgo.Session, *restful.Container, *db.ItemDBProvider, *db.PolicyDBProvider, *db.UserDBProvider) {
	s, err := mgo.Dial("localhost")
	if err != nil {
		Fail("could not setup db " + err.Error())
	}
	cont := restful.NewContainer()

	db.ReadPepper("/tmp/lsmsd_test_pepper")
	imgp := db.NewImageDBProvider(s, "lsmsd_test")

	itemp := db.NewItemDBProvider(s, "lsmsd_test", imgp)
	polp := db.NewPolicyDBProvider(s, "lsmsd_test")
	userp := db.NewUserDBProvider(s, itemp, polp, "lsmsd_test")
	us := webservice.NewUpdateService()
	auth := webservice.NewBasicAuthService(userp)
	iws := webservice.NewItemWebService(itemp, imgp, auth, us)
	pws := webservice.NewPolicyService(polp, auth, us)
	uws := webservice.NewUserService(userp, auth)
	cont.Add(iws.S)
	cont.Add(pws.S)
	cont.Add(uws.S)
	return s, cont, itemp, polp, userp
}

func populateDB(itm *db.ItemDBProvider, pol *db.PolicyDBProvider, usr *db.UserDBProvider) {
	populateItemDB(itm)
	populatePolicyDB(pol)
	populateUserDB(usr)
}

func populateItemDB(itm *db.ItemDBProvider) {
	var lastid uint64
	for i := 0; i != 10; i++ {
		i := db.Item{
			Name:        "test" + strconv.Itoa(i),
			Description: "testdesc",
			Contains:    []uint64{lastid},
			Owner:       "testuser",
			Maintainer:  "testuser",
			Usage:       "testpolicy",
			Discard:     "testpolicy",
		}

		var err error
		lastid, err = itm.CreateItem(&i)
		if err != nil {
			Fail("could not populate item db: " + err.Error())
		}
	}
}

func populatePolicyDB(pol *db.PolicyDBProvider) {
	for i := 0; i != 10; i++ {
		p := db.Policy{Name: strconv.Itoa(i), Description: "testdescr"}
		err := pol.CreatePolicy(&p)
		if err != nil {
			Fail("could not populate policy db: " + err.Error())
		}
	}
}

func populateUserDB(usr *db.UserDBProvider) {
	for i := 0; i != 10; i++ {
		sec := new(db.Secret)
		err := sec.SetPassword("testpw")
		if err != nil {
			Fail("could not populate user db: " + err.Error())
		}
		u := db.User{
			Name:   strconv.Itoa(i),
			EMail:  "test" + strconv.Itoa(i) + "@example.com",
			Secret: *sec,
		}
		err = usr.CreateUser(&u)
		if err != nil {
			Fail("could not populate user db: " + err.Error())
		}
	}
}

func flushDB(s *mgo.Session, itm *db.ItemDBProvider) {
	itm.Stop()
	coll, err := s.DB("lsmsd_test").CollectionNames()
	if err != nil {
		Fail("failed to clean up: " + err.Error())
	}
	for i := 0; i != len(coll); i++ {
		if strings.Contains(coll[i], "system") {
			continue
		}
		err := s.DB("lsmsd_test").C(coll[i]).DropCollection()
		if err != nil {
			Fail("failed to clean up: " + err.Error())
		}
	}
	s.Close()
}
