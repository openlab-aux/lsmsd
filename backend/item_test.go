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
	"encoding/json"
	"github.com/emicklei/go-restful"
	dmp "github.com/sergi/go-diff/diffmatchpatch"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

func Test_GetItemByIdReturns400InvalidID(t *testing.T) {
	cont := newTestContainer()
	s, err := startTestDB()
	if err != nil {
		t.Error("failed" + err.Error())
	}
	defer flushAndCloseTestDB(s, t)

	httpRequest, _ := http.NewRequest("GET", "/item/232323", nil)
	httpRequest.Header.Set("Content-Type", restful.MIME_JSON)
	httpWriter := httptest.NewRecorder()

	cont.ServeHTTP(httpWriter, httpRequest)
	if httpWriter.Code != http.StatusNotFound {
		t.Error("failed expected:" + strconv.Itoa(http.StatusNotFound) + "\n Got:" + strconv.Itoa(httpWriter.Code))
	}
}

func Test_GetItemByIdReturns500BadRequest(t *testing.T) {
	cont := newTestContainer()
	s, err := startTestDB()
	if err != nil {
		t.Error("failed: " + err.Error())
	}
	defer flushAndCloseTestDB(s, t)

	req, _ := http.NewRequest("GET", "/item/ſḥịŧ", nil)
	req.Header.Set("Content-Type", restful.MIME_JSON)
	hw := httptest.NewRecorder()

	cont.ServeHTTP(hw, req)
	if hw.Code != http.StatusBadRequest {
		t.Error("failed. expected:" + strconv.Itoa(http.StatusBadRequest) + "\n Got:" + strconv.Itoa(hw.Code))
	}
}

func TestGetItemById(t *testing.T) {
	cont := newTestContainer()
	s, err := startTestDB()
	if err != nil {
		t.Error("failed: " + err.Error())
	}
	defer flushAndCloseTestDB(s, t)
	itm := Item{EID: 23, Name: "testitem", Description: "testdesc", Contains: []uint64{2, 3}, Owner: "testuser", Maintainer: "test", Usage: "test", Discard: "disc"}
	id, err := createItem(&itm)
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
}

func Test_uint64Contains(t *testing.T) {
	A := []uint64{1, 2}
	B := []uint64{1, 3}
	if !uint64Contains(A, B[0]) || !uint64Contains(B, A[0]) {
		t.Error("failed")
	}
	if uint64Contains(A, B[1]) || uint64Contains(B, A[1]) {
		t.Error("failed")
	}
}

func Test_uint64Diff(t *testing.T) {
	A, B := []uint64{1, 2}, []uint64{1, 3}
	C := uint64Diff(A, B)
	_, ok := C["1"]
	if ok {
		t.Error("failed")
	}
	if C["2"] != dmp.DiffDelete {
		t.Error("failed")
	}
	if C["3"] != dmp.DiffInsert {
		t.Error("failed")
	}
}

func Benchmark_uint64Contains(b *testing.B) {
	A := make([]uint64, 10000)
	for i := 0; i != len(A); i++ {
		A[i] = uint64(rand.Int63())
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		uint64Contains(A, uint64(rand.Int63()))
	}
}
