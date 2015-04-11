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
	dmp "github.com/sergi/go-diff/diffmatchpatch"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"time"
)

type PolicyDBProvider struct {
	c  *mgo.Collection
	ch *mgo.Collection
}

func NewPolicyDBProvider(s *mgo.Session, dbname string) *PolicyDBProvider {
	res := new(PolicyDBProvider)
	res.c = s.DB(dbname).C("policy")
	res.ch = s.DB(dbname).C("policy_history")
	return res
}

type Policy struct {
	ID          bson.ObjectId `bson:"_id,omitempty" json:"-"`
	Name        string
	Description string
}

type PolicyHistory struct {
	ID        bson.ObjectId `bson:"_id,omitempty" json:"-"`
	Timestamp time.Time     `bson:"-" json:",omitempty"`
	User      string
	Policy    map[string]interface{}
}

func (p *Policy) NewPolicyHistory(po *Policy, user string) *PolicyHistory {
	res := new(PolicyHistory)
	res.Policy = make(map[string]interface{})
	res.User = user
	res.Policy["name"] = p.Name

	if po == nil {
		res.Policy["deleted"] = true
		return res
	}
	if p.Description != po.Description {
		d := dmp.New()
		d.DiffTimeout = 200 * time.Millisecond
		res.Policy["description"] = d.DiffMain(p.Description, po.Description, true)
	}
	return res
}

func (p *PolicyDBProvider) GetPolicyByName(name string) (Policy, error) {
	res := Policy{}
	err := p.c.Find(bson.M{"name": name}).One(&res)
	return res, err
}

func (p *PolicyDBProvider) GetPolicyLog(name string) ([]PolicyHistory, error) {
	res := make([]PolicyHistory, 0)
	err := p.ch.Find(bson.M{"policy.name": name}).All(&res)
	for i := 0; i != len(res); i++ {
		res[i].Timestamp = res[i].ID.Time()
	}
	return res, err
}

func (p *PolicyDBProvider) GetPolicyLogByUsername(name string) (*[]PolicyHistory, error) {
	ph := make([]PolicyHistory, 0)
	err := p.ch.Find(bson.M{"user": name}).All(&ph)
	return &ph, err
}

func (p *PolicyDBProvider) ListPolicy() ([]Policy, error) {
	pol := make([]Policy, 0)
	err := p.c.Find(nil).All(&pol)
	return pol, err
}

func (p *PolicyDBProvider) UpdatePolicy(pol *Policy, ph *PolicyHistory) error {
	err := p.ch.Insert(ph)
	if err != nil {
		return err
	}
	return p.c.Update(bson.M{"name": pol.Name}, pol)
}

func (p *PolicyDBProvider) CheckPolicyExistance(pol *Policy) bool {
	temp := Policy{}
	err := p.c.Find(bson.M{"name": pol.Name}).One(&temp)
	if err != nil {
		if err.Error() == "not found" {
			return false
		} else {
			log.WithFields(log.Fields{"Err": err}).
				Debug("TODO: check more error codes here")
			return true
		}
	}
	return true
}

func (p *PolicyDBProvider) CreatePolicy(pol *Policy) error {
	return p.c.Insert(pol)
}

func (p *PolicyDBProvider) DeletePolicy(pol *Policy, ph *PolicyHistory) error {
	err := p.ch.Insert(ph)
	if err != nil {
		return err
	}
	return p.c.Remove(bson.M{"name": pol.Name})
}
