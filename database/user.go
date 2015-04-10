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
	log "github.com/Sirupsen/logrus"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type UserDBProvider struct {
	c *mgo.Collection
	i *ItemDBProvider
	p *PolicyDBProvider
}

func NewUserDBProvider(s *mgo.Session, i *ItemDBProvider, p *PolicyDBProvider, dbname string) *UserDBProvider {
	res := new(UserDBProvider)
	res.c = s.DB(dbname).C("user")
	res.i = i
	res.p = p
	return res
}

type User struct {
	ID       bson.ObjectId `bson:"_id,omitempty" json:"-"`
	Name     string        `description:"The unique identifier of a user. Let your user know that they should choose it wisely"`
	EMail    string
	Password string `bson:"-" json:",omitempty" description:"Use this field to set a new password. This field will never occour in responses."`

	Secret Secret `json:"-"`
}

type UserActionHistory struct {
	ItemChanges   []ItemHistory
	PolicyChanges []PolicyHistory
}

func (p *UserDBProvider) GetUserByName(name string) (User, error) {
	res := User{}
	err := p.c.Find(bson.M{"name": name}).One(&res)
	return res, err
}

func (p *UserDBProvider) GetUserLogByName(name string) (*UserActionHistory, error) {
	ih, err := p.i.GetItemLogByUsername(name)
	if err != nil {
		return nil, err
	}
	ph, err := p.p.GetPolicyLogByUsername(name)
	if err != nil {
		return nil, err
	}
	ul := new(UserActionHistory)
	ul.ItemChanges = *ih
	ul.PolicyChanges = *ph
	return ul, nil
}

func (p *UserDBProvider) ListUser() ([]User, error) {
	usr := make([]User, 0)
	err := p.c.Find(nil).All(&usr)
	return usr, err
}

func (p *UserDBProvider) UpdateUser(usr *User) error {
	return p.c.Update(bson.M{"name": usr.Name}, usr)
}

func (p *UserDBProvider) CreateUser(usr *User) error {
	return p.c.Insert(usr)
}

func (p *UserDBProvider) CheckUserExistance(usr *User) bool {
	temp := User{}
	err := p.c.Find(bson.M{"name": usr.Name}).One(&temp)
	if err != nil {
		if err.Error() == "not found" {
			return false
		} else {
			log.WithFields(log.Fields{"Err": err}).
				Debug("TODO: check more error codes here")
			return true // if something strange in the neighbourhood …
			// what you gonna do? - do not register a new user!
		}
	}
	return true
}

func (p *UserDBProvider) DeleteUser(name string) error {
	return p.c.Remove(bson.M{"name": name})
}
