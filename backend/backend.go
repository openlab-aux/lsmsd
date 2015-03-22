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
)

const (
	ERROR_INVALID_ID    = "Error: Invalid ID"
	ERROR_STMT_PREPARE  = "Error: Statement prepare failed"
	ERROR_INVALID_INPUT = "ERROR: Invalid Input"
	ERROR_INTERNAL      = "Error: Internal Server Error"
	ERROR_INSERT        = "Error: DB Insert failed"
	ERROR_QUERY         = "Error: DB Query failed"
)

var db *mgo.Session

var uCol *mgo.Collection
var iCol *mgo.Collection
var pCol *mgo.Collection

var idgen *idgenerator

func RegisterDatabase(s *mgo.Session, dbname string) {
	db = s
	uCol = s.DB(dbname).C("user")
	iCol = s.DB(dbname).C("item")
	pCol = s.DB(dbname).C("policy")
	idgen = NewIDGenerator(s.DB(dbname).C("counters"))
}

func CloseIDGen() {
	idgen.StopIDGenerator()
}
