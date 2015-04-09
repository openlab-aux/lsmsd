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
	"gopkg.in/mgo.v2"
	"testing"
)

func startTestDB() (*mgo.Session, error) {
	s, err := mgo.Dial("localhost")
	if err != nil {
		return nil, err
	}
	RegisterDatabase(s, "lsmsd_test", &Mailconfig{})
	return s, nil
}

func flushAndCloseTestDB(m *mgo.Session, t *testing.T) {
	err := uCol.DropCollection()
	if err != nil && err.Error() != "ns not found" {
		t.Error("failed to clean up: " + err.Error())
	}
	err = iCol.DropCollection()
	if err != nil && err.Error() != "ns not found" {
		t.Error("failed to clean up: " + err.Error())
	}
	err = ihCol.DropCollection()
	if err != nil && err.Error() != "ns not found" {
		t.Error("failed to clean up: " + err.Error())
	}
	err = pCol.DropCollection()
	if err != nil && err.Error() != "ns not found" {
		t.Error("failed to clean up: " + err.Error())
	}
	err = phCol.DropCollection()
	if err != nil && err.Error() != "ns not found" {
		t.Error("failed to clean up: " + err.Error())
	}
	idgen.StopIDGenerator()
	m.Close()
}
