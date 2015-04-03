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
	"crypto/rand"
	log "github.com/Sirupsen/logrus"
	"gopkg.in/mgo.v2"
	"os"
)

const (
	ERROR_INVALID_ID    = "Error: Invalid ID"
	ERROR_STMT_PREPARE  = "Error: Statement prepare failed"
	ERROR_INVALID_INPUT = "ERROR: Invalid Input"
	ERROR_INTERNAL      = "Error: Internal Server Error"
	ERROR_INSERT        = "Error: DB Insert failed"
	ERROR_QUERY         = "Error: DB Query failed"
	PEPPER_SIZE         = 64
)

var db *mgo.Session

var uCol *mgo.Collection
var iCol *mgo.Collection
var ihCol *mgo.Collection
var pCol *mgo.Collection
var phCol *mgo.Collection

var pepper []byte

var idgen *idgenerator

func RegisterDatabase(s *mgo.Session, dbname string) {
	db = s
	uCol = s.DB(dbname).C("user")
	iCol = s.DB(dbname).C("item")
	ihCol = s.DB(dbname).C("item_history")
	pCol = s.DB(dbname).C("policy")
	phCol = s.DB(dbname).C("policy_history")
	idgen = NewIDGenerator(s.DB(dbname).C("counters"))
}

func ReadPepper(path string) {
	f, er := os.Open(path)
	if er != nil {
		err := er.(*os.PathError)
		log.WithFields(log.Fields{"Path": err.Path, "Op": err.Op}).Debug(err.Err)
		if err.Err.Error() == "no such file or directory" {
			log.Warn("Pepper file not found - creating ...")
			pepper = createPepper(path)
			return
		}
		log.Fatal(err)
	}
	defer f.Close()
	fi, er := f.Stat()
	if er != nil {
		log.Fatal(er)
	}
	if fi.Size() != PEPPER_SIZE {
		log.WithFields(log.Fields{"File Size": fi.Size(), "Expected Size": PEPPER_SIZE}).Fatal("Invalid pepper length - your file may be corrupt. Check your disk for errors.")
	}
	pepper = make([]byte, PEPPER_SIZE)
	bytes, er := f.Read(pepper)
	if er != nil || bytes != PEPPER_SIZE {
		log.WithFields(log.Fields{"Read": bytes, "Expected": PEPPER_SIZE}).Fatal(er)
	}
}

func createPepper(path string) []byte {
	res := make([]byte, PEPPER_SIZE)
	b, err := rand.Read(res)
	if err != nil || b != PEPPER_SIZE {
		log.WithFields(log.Fields{"Read": b, "Expected": PEPPER_SIZE}).Fatal(err)
	}
	f, err := os.Create(path)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	err = f.Chmod(0600)
	if err != nil {
		log.Fatal(err)
	}
	b, err = f.Write(res)
	if err != nil || b != PEPPER_SIZE {
		log.Fatal(err)
	}
	return res
}

func CloseIDGen() {
	idgen.StopIDGenerator()
}
