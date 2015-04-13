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
	//. "github.com/openlab-aux/lsmsd/webservice"
	"bytes"
	"encoding/json"
	"github.com/emicklei/go-restful"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	db "github.com/openlab-aux/lsmsd/database"
	"gopkg.in/mgo.v2"
	"net/http"
	"net/http/httptest"
)

var _ = Describe("User", func() {
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

	Describe("Retrieve a user from the database", func() {
		JustBeforeEach(func() {
			rd := bytes.NewReader(body)
			req, _ = http.NewRequest("GET", "/user/0", rd)
			req.Header.Set("Content-Type", restful.MIME_JSON)
		})

		Context("with invalid identifier", func() {
			It("should return a 404 not found", func() {
				cont.ServeHTTP(hw, req)

				Expect(hw.Code).To(Equal(http.StatusNotFound))
			})
		})

		Context("with valid id", func() {
			BeforeEach(func() {
				populateUserDB(usr)
			})

			It("should return a 200 OK", func() {
				cont.ServeHTTP(hw, req)

				Expect(hw.Code).To(Equal(http.StatusOK))
			})
		})
	})

	Describe("Delete a user from the database", func() {
		BeforeEach(func() {
			populateUserDB(usr)
			rd := bytes.NewReader(body)
			req, _ = http.NewRequest("DELETE", "/user/0", rd)
			req.Header.Set("Content-Type", restful.MIME_JSON)
		})

		Context("being authenticated", func() {
			Context("without being the related user", func() {
				JustBeforeEach(func() {
					req.SetBasicAuth("1", "testpw")
				})

				It("should be forbidden", func() {
					cont.ServeHTTP(hw, req)

					Expect(hw.Code).To(Equal(http.StatusForbidden))
				})

			})

			Context("being the user", func() {
				JustBeforeEach(func() {
					req.SetBasicAuth("0", "testpw")
				})

				It("should be OK", func() {
					cont.ServeHTTP(hw, req)

					Expect(hw.Code).To(Equal(http.StatusOK))
				})

			})
		})
		Context("being unauthenticated", func() {
			It("should say that I am unauthorized", func() {
				cont.ServeHTTP(hw, req)

				Expect(hw.Code).To(Equal(http.StatusUnauthorized))
			})

		})

	})

	Describe("Get a users log", func() {
		BeforeEach(func() {
			rd := bytes.NewReader(body)
			req, _ = http.NewRequest("GET", "/user/0/log", rd)
			req.Header.Set("Content-Type", restful.MIME_JSON)
		})

		PContext("with invalid identifier", func() {
			It("Should return 404 not found", func() {
				cont.ServeHTTP(hw, req)

				Expect(hw.Code).To(Equal(http.StatusUnauthorized))
			})
		})

		Context("with valid identifier", func() {
			BeforeEach(func() {
				populateUserDB(usr)
			})

			It("should return 200 OK", func() {
				cont.ServeHTTP(hw, req)

				Expect(hw.Code).To(Equal(http.StatusOK))
			})
		})
	})

	Describe("List all user", func() {
		BeforeEach(func() {
			rd := bytes.NewReader(body)
			req, _ = http.NewRequest("GET", "/user", rd)
			req.Header.Set("Content-Type", restful.MIME_JSON)
			populateUserDB(usr)
		})

		It("Should return 200 OK", func() {
			cont.ServeHTTP(hw, req)

			Expect(hw.Code).To(Equal(http.StatusOK))
		})
	})

	Describe("Update a user", func() {
		JustBeforeEach(func() {
			rd := bytes.NewReader(body)
			req, _ = http.NewRequest("PUT", "/user", rd)
			req.Header.Set("Content-Type", restful.MIME_JSON)
			populateUserDB(usr)
		})

		Context("being unauthenticated", func() {
			It("Should return 401 Unauthorized", func() {
				cont.ServeHTTP(hw, req)
				Expect(hw.Code).To(Equal(http.StatusUnauthorized))
			})
		})

		Context("being auhtenticated", func() {
			JustBeforeEach(func() {
				req.SetBasicAuth("1", "testpw")
			})

			Context("without being the user", func() {
				BeforeEach(func() {
					test := db.User{Name: "0", EMail: "test@example.example.com"}
					body, _ = json.Marshal(test)
				})

				It("should be forbidden", func() {
					cont.ServeHTTP(hw, req)
					Expect(hw.Code).To(Equal(http.StatusForbidden))
				})
			})

			Context("being the user", func() {
				BeforeEach(func() {
					test := db.User{Name: "1", EMail: "test@example.example.com"}
					body, _ = json.Marshal(test)
				})

				Context("with valid json", func() {
					It("should return 200 OK", func() {
						cont.ServeHTTP(hw, req)
						Expect(hw.Code).To(Equal(http.StatusOK))
					})
				})

				Context("with invalid json", func() {
					BeforeEach(func() {
						body = body[3:len(body)]
					})

					It("should return 400 Bad Request", func() {
						cont.ServeHTTP(hw, req)
						Expect(hw.Code).To(Equal(http.StatusBadRequest))
					})
				})
			})
		})
	})

	Describe("Register a user", func() {
		JustBeforeEach(func() {
			rd := bytes.NewReader(body)
			req, _ = http.NewRequest("POST", "/user", rd)
			req.Header.Set("Content-Type", restful.MIME_JSON)
		})

		Context("with already taken username", func() {
			BeforeEach(func() {
				populateUserDB(usr)
				test := db.User{Name: "0", EMail: "someuser@example.com", Password: "testpw1"}
				body, _ = json.Marshal(test)
			})

			It("should be forbidden", func() {
				cont.ServeHTTP(hw, req)
				Expect(hw.Code).To(Equal(http.StatusForbidden))
			})
		})

		Context("with usable username", func() {
			var (
				user *db.User
			)

			BeforeEach(func() {
				user = &db.User{Name: "0", EMail: "someuser@example.com"}
			})
			Context("without password", func() {
				BeforeEach(func() {
					body, _ = json.Marshal(user)
				})

				It("should be a bad request", func() {
					cont.ServeHTTP(hw, req)

					Expect(hw.Code).To(Equal(http.StatusBadRequest))
				})
			})

			Context("with password", func() {
				BeforeEach(func() {
					user.Password = "testpw1"
					body, _ = json.Marshal(user)
				})

				It("should be ok", func() {
					cont.ServeHTTP(hw, req)

					Expect(hw.Code).To(Equal(http.StatusOK))
				})
			})
		})
	})
})
