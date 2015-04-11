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

package database

import (
	"crypto/rand"
	"crypto/sha512"
	"errors"
	log "github.com/Sirupsen/logrus"
	"os"
)

const (
	pepperSize = 64
)

var pepper []byte

func ReadPepper(path string) {
	if path == "" {
		log.Panic("Pepperpath empty")
	}
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
	if fi.Size() != pepperSize {
		log.WithFields(log.Fields{"File Size": fi.Size(), "Expected Size": pepperSize}).Fatal("Invalid pepper length - your file may be corrupt. Check your disk for errors.")
	}
	pepper = make([]byte, pepperSize)
	bytes, er := f.Read(pepper)
	if er != nil || bytes != pepperSize {
		log.WithFields(log.Fields{"Read": bytes, "Expected": pepperSize}).Fatal(er)
	}
}

func createPepper(path string) []byte {
	res := make([]byte, pepperSize)
	b, err := rand.Read(res)
	if err != nil || b != pepperSize {
		log.WithFields(log.Fields{"Read": b, "Expected": pepperSize}).Fatal(err)
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
	if err != nil || b != pepperSize {
		log.Fatal(err)
	}
	return res
}

type Secret struct {
	Password [sha512.Size]byte `json:"-"`
	Salt     [64]byte          `json:"-"`
}

func (s *Secret) VerifyPassword(pw string) bool {
	input := s.assemblePassword(pw)
	for i := 0; i != len(s.Password); i++ {
		if s.Password[i] != input[i] {
			return false
		}
	}
	return true
}

func (s *Secret) SetPassword(pw string) error {
	err := s.genSalt()
	if err != nil {
		return err
	}

	s.Password = s.assemblePassword(pw)
	return nil
}

func (s *Secret) assemblePassword(pw string) [sha512.Size]byte {
	temp := make([]byte, len(pw)+len(s.Salt)+len(pepper))
	for i := 0; i != len(pw); i++ {
		temp[i] = pw[i]
	}
	for i := 0; i != len(s.Salt); i++ {
		temp[i+len(pw)] = s.Salt[i]
	}
	for i := 0; i != len(pepper); i++ {
		temp[i+len(pw)+len(s.Salt)] = pepper[i]
	}
	return sha512.Sum512(temp)
}

func (s *Secret) genSalt() error {
	temp := make([]byte, len(s.Salt))
	b, err := rand.Read(temp)
	if err != nil {
		return err
	}
	if b != len(s.Salt) {
		return errors.New("Read less than expected bytes")
	}
	for i := 0; i != len(s.Salt); i++ {
		s.Salt[i] = temp[i]
	}
	return nil
}
