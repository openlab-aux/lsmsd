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

package webservice

import (
	//	"encoding/json"
	"github.com/emicklei/go-restful"
	db "github.com/openlab-aux/lsmsd/database"
	//	dmp "github.com/sergi/go-diff/diffmatchpatch"
	"gopkg.in/mgo.v2"
	//	"math/rand"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

func newTestContainer(t testing.TB) (*mgo.Session, *restful.Container, *db.ItemDBProvider) {
	s, err := mgo.Dial("localhost")
	if err != nil {
		t.Error(err.Error())
	}
	cont := restful.NewContainer()

	itemp := db.NewItemDBProvider(s, "lsmsd_test")
	polp := db.NewPolicyDBProvider(s, "lsmsd_test")
	userp := db.NewUserDBProvider(s, itemp, polp, "lsmsd_test")
	auth := NewBasicAuthService(userp)
	iws := NewItemWebService(itemp, auth)
	pws := NewPolicyService(polp, auth)
	uws := NewUserService(userp, auth)
	cont.Add(iws.S)
	cont.Add(pws.S)
	cont.Add(uws.S)
	return s, cont, itemp
}

func flushDB(t testing.TB, s *mgo.Session, itm *db.ItemDBProvider) {
	itm.Stop()
	err := s.DB("lsmsd_test").DropDatabase()
	if err != nil {
		t.Error("failed to clean up: " + err.Error())
	}
	s.Close()
}

func Test_GetItemByIdReturns400InvalidID(t *testing.T) {
	s, cont, it := newTestContainer(t)
	defer flushDB(t, s, it)
	httpRequest, _ := http.NewRequest("GET", "/item/232323", nil)
	httpRequest.Header.Set("Content-Type", restful.MIME_JSON)
	httpWriter := httptest.NewRecorder()

	cont.ServeHTTP(httpWriter, httpRequest)
	if httpWriter.Code != http.StatusNotFound {
		t.Error("failed expected:" + strconv.Itoa(http.StatusNotFound) + "\n Got:" + strconv.Itoa(httpWriter.Code))
	}
}

func Test_GetItemByIdReturns500BadRequest(t *testing.T) {
	s, cont, it := newTestContainer(t)
	defer flushDB(t, s, it)

	req, _ := http.NewRequest("GET", "/item/ſḥịŧ", nil)
	req.Header.Set("Content-Type", restful.MIME_JSON)
	hw := httptest.NewRecorder()

	cont.ServeHTTP(hw, req)
	if hw.Code != http.StatusBadRequest {
		t.Error("failed. expected:" + strconv.Itoa(http.StatusBadRequest) + "\n Got:" + strconv.Itoa(hw.Code))
	}
}

/*func TestGetItemById(t *testing.T) {
	s, cont := newTestContainer(t)
	defer flushDB(t, s)
	itm := db.Item{EID: 23, Name: "testitem", Description: "testdesc", Contains: []uint64{2, 3}, Owner: "testuser", Maintainer: "test", Usage: "test", Discard: "disc"}
	id, err := reateItem(&itm)
	if err != nil {
		t.Error("failed: " + err.Error())
	}

	req, _ := http.NewRequest("GET", "/item/"+strconv.FormatUint(id, 10), nil)
	req.Header.Set("Content-Type", restful.MIME_JSON)
	hw := httptest.NewRecorder()

	cont.ServeHTTP(hw, req)
	if hw.Code != http.StatusOK {
		t.Error("failed. expected:" + strconv.Itoa(http.StatusOK) + "\n Got:" + strconv.Itoa(hw.Code))
	}
	body := hw.Body.Bytes()
	jitm := new(Item)
	err = json.Unmarshal(body, jitm)
	if err != nil {
		t.Error("failed: " + err.Error())
	}
	if itm.EID != jitm.EID {
		t.Error("id injection")
	}
	if itm.Name != jitm.Name {
		t.Error()
	}
	if itm.Description != jitm.Description {
		t.Error()
	}
	if len(itm.Contains) != len(jitm.Contains) {
		t.Error()
	}
	for i := 0; i != len(itm.Contains); i++ {
		if itm.Contains[i] != jitm.Contains[i] {
			t.Error()
		}
	}
	if itm.Owner != jitm.Owner {
		t.Error()
	}
	if itm.Maintainer != jitm.Maintainer {
		t.Error()
	}
	if itm.Usage != jitm.Usage {
		t.Error()
	}
	if itm.Discard != jitm.Discard {
		t.Error()
	}
}*/
