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
	log "github.com/Sirupsen/logrus"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"sync"
	"time"
)

type idgenerator struct {
	status chan int         // status channel, 1 triggers an exit
	getID  chan chan uint64 // result channel channel lol
	wg     sync.WaitGroup
	c      *mgo.Collection
}

type counter struct {
	ID    bson.ObjectId `bson:"_id,omitempty"`
	Type_ string
	Count uint64
}

// NewIDGenerator starts a new goroutine, which generates new autoincrementing
// IDs. Do not generate multiple instances of this object!
func NewIDGenerator(c *mgo.Collection) *idgenerator {
	res := idgenerator{make(chan int), make(chan chan uint64), *new(sync.WaitGroup), c}
	go res.generateID()
	res.wg.Add(1)
	return &res
}

func (i *idgenerator) StopIDGenerator() {
	i.status <- 1
	i.wg.Wait()
}

func (i *idgenerator) ResetCounter() {
	err := i.c.DropCollection()
	if err != nil && err.Error() != "ns not found" {
		log.Panic(err.Error())
	}
}

// GenerateID is a blocking function to get a incrementing id
func (i *idgenerator) GenerateID() uint64 {
	res := make(chan uint64)
	i.getID <- res
	return <-res
}

func (i *idgenerator) generateID() {
	defer i.wg.Done()
	var x counter
	//check if document exist
	err := i.c.Find(bson.M{"type_": "item"}).One(&x)
	if err != nil {
		if err.Error() != "not found" {
			log.Panic(err)
		} else {
			log.Info("Empty Database found.")
			i.c.Insert(&counter{"", "item", 0})
		}
	}

	hit := false

	for {
		select {
		case _ = <-i.status:
			return
		default:
			select {
			case ret := <-i.getID:
				hit = true
				var temp counter
				err := i.c.Find(bson.M{"type_": "item"}).One(&temp)
				if err != nil {
					log.Panic(err)
				}
				temp.Count++
				err = i.c.Update(bson.M{"type_": "item"}, &temp)
				if err != nil {
					log.Panic(err)
				}
				ret <- temp.Count

			default:
			}

		}
		if !hit { // Process all request every 75 Millisecond to save CPU time
			time.Sleep(75 * time.Millisecond)
		} else {
			hit = false
		}

	}
}
