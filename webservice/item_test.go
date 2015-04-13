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
	"encoding/json"
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

	Describe("Delete a item", func() {
		BeforeEach(func() {
			populateUserDB(usr)
		})

		Context("with valid id", func() {
			BeforeEach(func() {
				rd := bytes.NewReader(body)
				req, _ = http.NewRequest("DELETE", "/item/1", rd)
				req.Header.Set("Content-Type", restful.MIME_JSON)
			})
			Context("being authenticated", func() {
				JustBeforeEach(func() {
					req.SetBasicAuth("1", "testpw")
				})

				Context("but the item does not exist", func() {
					It("should return 404 not found", func() {
						cont.ServeHTTP(hw, req)

						Expect(hw.Code).To(Equal(http.StatusNotFound))
					})
				})

				Context("and the item does exist", func() {
					BeforeEach(func() {
						populateItemDB(itm)
					})
					It("should be ok", func() {
						cont.ServeHTTP(hw, req)

						Expect(hw.Code).To(Equal(http.StatusOK))
					})
				})
			})

			Context("being unauthenticated", func() {
				Context("and the item does not exist", func() {
					It("should return 401 unauthorized", func() {
						cont.ServeHTTP(hw, req)

						Expect(hw.Code).To(Equal(http.StatusUnauthorized))
					})
				})

				Context("and the item does exist", func() {
					It("should return 401 unauthorized", func() {
						cont.ServeHTTP(hw, req)

						Expect(hw.Code).To(Equal(http.StatusUnauthorized))
					})
				})
			})
		})

		Context("with invalid id", func() {
			BeforeEach(func() {
				rd := bytes.NewReader(body)
				req, _ = http.NewRequest("DELETE", "/item/ſħịŧ", rd)
				req.Header.Set("Content-Type", restful.MIME_JSON)
			})

			Context("being authenticated", func() {
				JustBeforeEach(func() {
					req.SetBasicAuth("1", "testpw")
				})

				It("should return 400 bad request", func() {
					cont.ServeHTTP(hw, req)

					Expect(hw.Code).To(Equal(http.StatusBadRequest))
				})
			})

			Context("being unauthenticated", func() {
				It("should return 401 unauthorized", func() {
					cont.ServeHTTP(hw, req)

					Expect(hw.Code).To(Equal(http.StatusUnauthorized))
				})
			})
		})
	})

	Describe("Hear a joke", func() {
		BeforeEach(func() {
			rd := bytes.NewReader(body)
			req, _ = http.NewRequest("GET", "/item/coffee", rd)
			req.Header.Set("Content-Type", restful.MIME_JSON)
		})

		Context("a really bad one", func() {
			It("should tell you, that I am a teapot", func() {
				cont.ServeHTTP(hw, req)
				Expect(hw.Code).To(Equal(http.StatusTeapot))
			})
		})
	})

	Describe("Get the history of a item", func() {
		Context("with invalid identifier", func() {
			BeforeEach(func() {
				rd := bytes.NewReader(body)
				req, _ = http.NewRequest("GET", "/item/ſħịŧ/log", rd)
				req.Header.Set("Content-Type", restful.MIME_JSON)
			})

			It("should return 400 bad request", func() {
				cont.ServeHTTP(hw, req)

				Expect(hw.Code).To(Equal(http.StatusBadRequest))
			})
		})

		Context("with valid identifier", func() {
			BeforeEach(func() {
				rd := bytes.NewReader(body)
				req, _ = http.NewRequest("GET", "/item/1/log", rd)
				req.Header.Set("Content-Type", restful.MIME_JSON)
			})
			PContext("which does not exist", func() {
				It("Should return 404 not found", func() {
					cont.ServeHTTP(hw, req)

					Expect(hw.Code).To(Equal(http.StatusNotFound))
				})
			})
			Context("which does exist", func() {
				BeforeEach(func() {
					populateItemDB(itm)
				})

				It("Should be ok", func() {
					cont.ServeHTTP(hw, req)

					Expect(hw.Code).To(Equal(http.StatusOK))
				})
			})
		})
	})

	Describe("List all items", func() {
		BeforeEach(func() {
			rd := bytes.NewReader(body)
			req, _ = http.NewRequest("GET", "/item", rd)
			req.Header.Set("Content-Type", restful.MIME_JSON)
			populateItemDB(itm)
		})

		It("Should be ok", func() {
			cont.ServeHTTP(hw, req)

			Expect(hw.Code).To(Equal(http.StatusOK))
		})
	})

	Describe("Update a item", func() {
		JustBeforeEach(func() {
			rd := bytes.NewReader(body)
			req, _ = http.NewRequest("PUT", "/item", rd)
			req.Header.Set("Content-Type", restful.MIME_JSON)
		})

		Context("being unauthenticated", func() {
			BeforeEach(func() {
				test := db.Item{EID: 1, Name: "New Name"}
				body, _ = json.Marshal(test)
			})

			It("should return 401 unauthorized", func() {
				cont.ServeHTTP(hw, req)
				Expect(hw.Code).To(Equal(http.StatusUnauthorized))
			})
		})

		Context("being authenticated", func() {
			JustBeforeEach(func() {
				populateUserDB(usr)
				req.SetBasicAuth("1", "testpw")
			})

			Context("but the item does not exist", func() {
				BeforeEach(func() {
					test := db.Item{EID: 1337, Name: "Does not exist"}
					body, _ = json.Marshal(test)
				})
				It("should return 404 not found", func() {
					cont.ServeHTTP(hw, req)
					Expect(hw.Code).To(Equal(http.StatusNotFound))
				})
			})

			Context("and the item does exist", func() {
				BeforeEach(func() {
					test := db.Item{EID: 1, Name: "Does exist"}
					body, _ = json.Marshal(test)
					populateItemDB(itm)
				})

				Context("but the json is invalid", func() {
					BeforeEach(func() {
						body = body[3:len(body)]
					})

					It("should return 400 bad request", func() {
						cont.ServeHTTP(hw, req)
						Expect(hw.Code).To(Equal(http.StatusBadRequest))
					})
				})

				Context("with valid json", func() {
					It("should be ok", func() {
						cont.ServeHTTP(hw, req)
						Expect(hw.Code).To(Equal(http.StatusOK))
					})
				})
			})
		})

	})

	Describe("Insert a item", func() {
		JustBeforeEach(func() {
			rd := bytes.NewReader(body)
			req, _ = http.NewRequest("POST", "/item", rd)
			req.Header.Set("Content-Type", restful.MIME_JSON)
			populateUserDB(usr)
		})

		Context("without authentication", func() {
			BeforeEach(func() {
				test := db.Item{Name: "new test item", Description: "#Awesome"}
				body, _ = json.Marshal(test)
			})

			It("should return 401 unauthorized", func() {
				cont.ServeHTTP(hw, req)
				Expect(hw.Code).To(Equal(http.StatusUnauthorized))
			})

		})

		Context("with authentication", func() {
			BeforeEach(func() {
				test := db.Item{Name: "new test item", Description: "#Awesome"}
				body, _ = json.Marshal(test)
			})

			JustBeforeEach(func() {
				req.SetBasicAuth("1", "testpw")
			})

			Context("with valid json", func() {
				It("should be ok", func() {
					cont.ServeHTTP(hw, req)
					Expect(hw.Code).To(Equal(http.StatusOK))
				})

			})

			Context("with invalid json", func() {
				BeforeEach(func() {
					body = body[3:len(body)]
				})

				It("should return 400 bad request", func() {
					cont.ServeHTTP(hw, req)
					Expect(hw.Code).To(Equal(http.StatusBadRequest))
				})

			})

		})
	})
})
