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
	"github.com/emicklei/go-restful"
	db "github.com/openlab-aux/lsmsd/database"
	//	. "github.com/openlab-aux/lsmsd/webservice"
	"bytes"
	//	"encoding/json"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gopkg.in/mgo.v2"
	"net/http"
	"net/http/httptest"
)

var _ = Describe("Item", func() {
	var (
		session *mgo.Session
		cont    *restful.Container
		itm     *db.ItemDBProvider
		pol     *db.PolicyDBProvider
		usr     *db.UserDBProvider
		hw      *httptest.ResponseRecorder
		req     *http.Request
		body    []byte
	)

	BeforeEach(func() {
		session, cont, itm, pol, usr = newTestContainer()
		hw = httptest.NewRecorder()
	})

	AfterEach(func() {
		flushDB(session, itm)
	})

	Describe("Retrieve a item from the database", func() {
		Context("With an invalid identifier", func() {
			JustBeforeEach(func() {
				rd := bytes.NewReader(body)
				req, _ = http.NewRequest("GET", "/item/ſhịŧ", rd)
				req.Header.Set("Content-Type", restful.MIME_JSON)
			})
			It("should be a bad request", func() {
				cont.ServeHTTP(hw, req)
				Expect(hw.Code).To(Equal(http.StatusBadRequest))
			})
		})

		Context("With valid identifier", func() {
			JustBeforeEach(func() {
				rd := bytes.NewReader(body)
				req, _ = http.NewRequest("GET", "/item/1", rd)
				req.Header.Set("Content-Type", restful.MIME_JSON)
			})
			Context("which does not exist", func() {
				It("should return 404 not found", func() {
					cont.ServeHTTP(hw, req)
					Expect(hw.Code).To(Equal(http.StatusNotFound))
				})
			})

			Context("which does exist", func() {
				BeforeEach(func() {
					populateItemDB(itm)
				})
				It("should return 200 ok", func() {
					cont.ServeHTTP(hw, req)
					Expect(hw.Code).To(Equal(http.StatusOK))
				})
			})
		})
	})
})
