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
	"crypto/sha512"
	"errors"
)

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
