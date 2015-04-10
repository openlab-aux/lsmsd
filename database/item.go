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
	"strconv"
	"time"
)

type ItemDBProvider struct {
	c     *mgo.Collection
	ch    *mgo.Collection
	idgen *idgenerator
}

func NewItemDBProvider(s *mgo.Session, dbname string) *ItemDBProvider {
	res := new(ItemDBProvider)
	res.c = s.DB(dbname).C("item")
	res.ch = s.DB(dbname).C("item_history")
	res.idgen = NewIDGenerator(s.DB(dbname).C("counters"))
	return res
}

func (p *ItemDBProvider) Stop() {
	p.idgen.StopIDGenerator()
}

func (p *ItemDBProvider) GetItemById(id uint64) (Item, error) {
	res := Item{}
	err := p.c.Find(bson.M{"eid": id}).One(&res)
	return res, err
}

func (p *ItemDBProvider) GetItemLog(id uint64) ([]ItemHistory, error) {
	res := make([]ItemHistory, 0)
	err := p.ch.Find(bson.M{"item.eid": id}).All(&res)
	for i := 0; i != len(res); i++ {
		res[i].Timestamp = res[i].ID.Time()
	}
	return res, err
}

func (p *ItemDBProvider) GetItemLogByUsername(name string) (*[]ItemHistory, error) {
	i := make([]ItemHistory, 0)
	err := p.ch.Find(bson.M{"user": name}).All(&i)
	return &i, err
}

func (p *ItemDBProvider) CreateItem(itm *Item) (uint64, error) {
	itm.EID = p.idgen.GenerateID()
	log.WithFields(log.Fields{"ID": itm.EID}).Debug("Generated ID")
	err := p.c.Insert(itm)
	return itm.EID, err
}

func (p *ItemDBProvider) ListItem() ([]Item, error) {
	itm := make([]Item, 0)
	err := p.c.Find(nil).All(&itm)
	return itm, err
}

func (p *ItemDBProvider) UpdateItem(itm *Item, ih *ItemHistory) error {
	err := p.ch.Insert(ih)
	if err != nil {
		return err
	}
	return p.c.Update(bson.M{"eid": itm.EID}, itm)
}

func (p *ItemDBProvider) CheckItemExistance(itm *Item) bool {
	temp := Item{}
	err := p.c.Find(bson.M{"eid": itm.EID}).One(&temp)
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

func (p *ItemDBProvider) DeleteItem(itm *Item, ih *ItemHistory) error {
	err := p.ch.Insert(ih)
	if err != nil {
		return err
	}
	return p.c.Remove(bson.M{"eid": itm.EID})
}

type Item struct {
	ID          bson.ObjectId `bson:"_id,omitempty" json:"-"`
	EID         uint64        `json:"Id"`
	Name        string        `bson:",omitempty"`
	Description string        `bson:",omitempty"`
	Contains    []uint64      `bson:",omitempty"`
	Owner       string        `bson:",omitempty"`
	Maintainer  string        `bson:",omitempty"`
	Usage       string        `bson:",omitempty"`
	Discard     string        `bson:",omitempty"`
}

type ItemHistory struct {
	ID        bson.ObjectId `bson:"_id,omitempty" json:"-"`
	Timestamp time.Time     `bson:"-" json:",omitempty"`
	User      string
	Item      map[string]interface{}
}

func (i *Item) NewItemHistory(it *Item, user string) *ItemHistory {
	res := new(ItemHistory)
	res.Item = make(map[string]interface{})
	res.Item["eid"] = i.EID
	res.User = user

	if it == nil {
		res.Item["deleted"] = true
		return res
	}
	//res.Item["_id"] = i.ID
	if i.Name != it.Name {
		res.Item["name"] = it.Name
	}
	if i.Description != it.Description {
		d := dmp.New()
		d.DiffTimeout = 200 * time.Millisecond
		res.Item["description"] = d.DiffMain(i.Description, it.Description, true)
	}
	condiff := uint64Diff(i.Contains, it.Contains)
	if len(condiff) > 0 {
		res.Item["contains"] = condiff
	}
	if i.Owner != it.Owner {
		res.Item["owner"] = it.Owner
	}
	if i.Maintainer != it.Maintainer {
		res.Item["maintainer"] = it.Maintainer
	}
	if i.Usage != it.Usage {
		res.Item["usage"] = it.Usage
	}
	if i.Discard != it.Discard {
		res.Item["discard"] = it.Discard
	}
	return res
}

func uint64Diff(u1, u2 []uint64) map[string]dmp.Operation {
	// mgo.bson does only support strings as keys
	res := make(map[string]dmp.Operation)
	for i := 0; i != len(u1); i++ {
		if !uint64Contains(u2, u1[i]) {
			res[strconv.FormatUint(u1[i], 10)] = dmp.DiffDelete
		}
	}
	for i := 0; i != len(u2); i++ {
		if !uint64Contains(u1, u2[i]) {
			res[strconv.FormatUint(u2[i], 10)] = dmp.DiffInsert
		}
	}
	return res
}

func uint64Contains(sl []uint64, u uint64) bool {
	for i := 0; i != len(sl); i++ {
		if sl[i] == u {
			return true
		}
	}
	return false
}
