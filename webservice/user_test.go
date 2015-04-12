package webservice_test

import (
	//. "github.com/openlab-aux/lsmsd/webservice"
	"bytes"
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
})
