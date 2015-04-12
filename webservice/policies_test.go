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

var _ = Describe("Policies", func() {
	var (
		session *mgo.Session
		cont    *restful.Container
		itm     *db.ItemDBProvider
		pol     *db.PolicyDBProvider
		usr     *db.UserDBProvider
		hw      *httptest.ResponseRecorder
	)

	BeforeEach(func() {
		session, cont, itm, pol, usr = newTestContainer()
		hw = httptest.NewRecorder()
	})

	AfterEach(func() {
		flushDB(session, itm)
	})

	Describe("Retrieve a policy from the database", func() {
		Context("Whith an invalid identifier", func() {
			It("Should return a 404", func() {
				req, _ := http.NewRequest("GET", "/policy/0", nil)
				req.Header.Set("Content-Type", restful.MIME_JSON)

				cont.ServeHTTP(hw, req)

				Expect(hw.Code).To(Equal(http.StatusNotFound))
			})
		})

		Context("With a valid identifier", func() {
			BeforeEach(func() {
				populatePolicyDB(pol)
			})

			It("Shoud return a 200", func() {
				req, _ := http.NewRequest("GET", "/policy/0", nil)
				req.Header.Set("Content-Type", restful.MIME_JSON)

				cont.ServeHTTP(hw, req)

				Expect(hw.Code).To(Equal(http.StatusOK))
			})
		})
	})

	Describe("Delete a policy from the database", func() {
		Context("Without authentication", func() {
			It("Should return 401 Unauthorized", func() {
				req, _ := http.NewRequest("DELETE", "/policy/0", nil)
				req.Header.Set("Content-Type", restful.MIME_JSON)

				cont.ServeHTTP(hw, req)

				Expect(hw.Code).To(Equal(http.StatusUnauthorized))
			})
		})

		Context("With authentication", func() {
			BeforeEach(func() {
				populatePolicyDB(pol)
				populateUserDB(usr)
			})
			It("Should return 200", func() {
				req, _ := http.NewRequest("DELETE", "/policy/0", nil)
				req.Header.Set("Content-Type", restful.MIME_JSON)
				req.SetBasicAuth("0", "testpw")

				cont.ServeHTTP(hw, req)

				Expect(hw.Code).To(Equal(http.StatusOK))
			})
			Context("but with invalid id", func() {
				It("Should return 404", func() {
					req, _ := http.NewRequest("DELETE", "/policy/INVALID", nil)
					req.Header.Set("Content-Type", restful.MIME_JSON)
					req.SetBasicAuth("0", "testpw")

					cont.ServeHTTP(hw, req)

					Expect(hw.Code).To(Equal(http.StatusNotFound))
				})
			})
		})

	})

	Describe("Get a policys log", func() {
		Context("With invalid identifier", func() {
			It("Should return 404 Not Found", func() {
				req, _ := http.NewRequest("GET", "/policy/INVALID/log", nil)
				req.Header.Set("Content-Type", restful.MIME_JSON)

				cont.ServeHTTP(hw, req)

				Expect(hw.Code).To(Equal(http.StatusNotFound))
			})
		})

		Context("With valid identifier", func() {
			BeforeEach(func() {
				populatePolicyDB(pol)
			})
			It("Should return 200 OK", func() {
				req, _ := http.NewRequest("GET", "/policy/0/log", nil)
				req.Header.Set("Content-Type", restful.MIME_JSON)

				cont.ServeHTTP(hw, req)

				Expect(hw.Code).To(Equal(http.StatusOK))
			})
		})
	})

	Describe("List all policies", func() {
		var (
			req *http.Request
		)

		BeforeEach(func() {
			req, _ = http.NewRequest("GET", "/policy", nil)
			req.Header.Set("Content-Type", restful.MIME_JSON)
		})

		Context("With empty DB", func() {
			It("Should return 200 OK", func() {
				cont.ServeHTTP(hw, req)

				Expect(hw.Code).To(Equal(http.StatusOK))
			})
		})

		Context("With filled DB", func() {
			BeforeEach(func() {
				populatePolicyDB(pol)
			})
			It("should return 200 OK", func() {
				cont.ServeHTTP(hw, req)

				Expect(hw.Code).To(Equal(http.StatusOK))
			})
		})
	})

	Describe("Insert a policy", func() {
		var (
			req  *http.Request
			body []byte
		)

		JustBeforeEach(func() {
			rd := bytes.NewReader(body)
			req, _ = http.NewRequest("POST", "/policy", rd)
			req.Header.Set("Content-Type", restful.MIME_JSON)
		})

		AfterEach(func() {
			body = nil
			req = nil
		})

		Context("Without authentication", func() {
			It("Should return 401 Unauthorized", func() {
				cont.ServeHTTP(hw, req)

				Expect(hw.Code).To(Equal(http.StatusUnauthorized))
			})
		})

		Context("With authentication", func() {
			JustBeforeEach(func() {
				populateUserDB(usr)
				req.SetBasicAuth("0", "testpw")
			})
			Context("Without body", func() {
				It("Should return 400 Bad Request", func() {
					cont.ServeHTTP(hw, req)
					Expect(hw.Code).To(Equal(http.StatusBadRequest))
				})
			})
			Context("and body", func() {
				BeforeEach(func() {
					test := db.Policy{Name: "testpolicy123", Description: "testdesc"}
					body, _ = json.Marshal(test)
				})

				It("Should return 200 OK", func() {
					cont.ServeHTTP(hw, req)
					Expect(hw.Code).To(Equal(http.StatusOK))
				})
			})
			Context("and broken body", func() {
				BeforeEach(func() {
					test := db.Policy{Name: "testpolicy123", Description: "testdesc"}
					b, _ := json.Marshal(test)
					body = b[3:len(b)]
				})

				It("Should return 400 Bad Request", func() {
					cont.ServeHTTP(hw, req)
					Expect(hw.Code).To(Equal(http.StatusBadRequest))
				})
			})
		})
	})

	Describe("Update a Policy", func() {
		var (
			req  *http.Request
			body []byte
		)

		JustBeforeEach(func() {
			rd := bytes.NewReader(body)
			req, _ = http.NewRequest("PUT", "/policy", rd)
			req.Header.Set("Content-Type", restful.MIME_JSON)
		})

		Context("Without authentication", func() {
			It("Should return 401 Unauthorized", func() {
				cont.ServeHTTP(hw, req)
				Expect(hw.Code).To(Equal(http.StatusUnauthorized))
			})
		})

		Context("With authentication", func() {
			JustBeforeEach(func() {
				populateUserDB(usr)
				populatePolicyDB(pol)
				req.SetBasicAuth("0", "testpw")
			})

			Context("Without body", func() {
				BeforeEach(func() {
					body = []byte{}
				})
				It("Should return 400 Bad Request", func() {
					cont.ServeHTTP(hw, req)
					Expect(hw.Code).To(Equal(http.StatusBadRequest))
				})
			})

			Context("and body", func() {
				BeforeEach(func() {
					test := db.Policy{Name: "0", Description: "testd."}
					body, _ = json.Marshal(test)
				})

				It("Should return 200 OK", func() {
					cont.ServeHTTP(hw, req)
					Expect(hw.Code).To(Equal(http.StatusOK))
				})
			})

			Context("and broken body", func() {
				BeforeEach(func() {
					test := db.Policy{Name: "0", Description: "testd."}
					b, _ := json.Marshal(test)
					body = b[3:len(b)]
				})

				It("Should return 400 Bad Request", func() {
					cont.ServeHTTP(hw, req)
					Expect(hw.Code).To(Equal(http.StatusBadRequest))
				})
			})
		})
	})
})
