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
	//	"database/sql"
	log "github.com/Sirupsen/logrus"
	"github.com/emicklei/go-restful"
	//"html/template"
	//	"github.com/fatih/structs"
	dmp "github.com/sergi/go-diff/diffmatchpatch"
	"gopkg.in/mgo.v2/bson"
	"net/http"
	"strconv"
	//"strings"
	"time"
)

type Item struct {
	ID          bson.ObjectId `bson:"_id,omitempty" json:"-"`
	EID         uint64        `json:"Id"`
	Name        string        `bson:",omitempty"`
	Description string        `bson:",omitempty" description:"This string should be in Github Flavored Markdown"`
	Contains    []uint64      `bson:",omitempty"`
	Owner       string        `bson:",omitempty"`
	Maintainer  string        `bson:",omitempty"`
	Usage       string        `bson:",omitempty"`
	Discard     string        `bson:",omitempty"`
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

type ItemHistory struct {
	ID        bson.ObjectId `bson:"_id,omitempty" json:"-"`
	Timestamp time.Time     `bson:"-" json:",omitempty"`
	User      string
	Item      map[string]interface{}
}

func NewItemService() *restful.WebService {
	service := new(restful.WebService)
	service.
		Path("/item").
		Doc("Item related services").
		ApiVersion("0.1").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	service.Route(service.GET("/{id}").
		Param(restful.PathParameter("id", "Item ID")).
		Doc("Returns a single item identified by its ID").
		//Returns(http.StatusOK, "Item request successful", Item{}).
		To(GetItemById).
		Writes(Item{}).
		Do(returnsInternalServerError, returnsNotFound, returnsBadRequest))

	service.Route(service.GET("/{id}/log").
		Param(restful.PathParameter("id", "Item ID")).
		Doc("Returns the items changelog").
		To(GetItemLog).
		Writes(ItemHistory{}).
		Do(returnsInternalServerError, returnsNotFound, returnsBadRequest))

	service.Route(service.GET("").
		Doc("List all available items (this may be replaced by a paginated version)").
		To(ListItem).
		Writes([]Item{}).
		Do(returnsInternalServerError))

	service.Route(service.PUT("").
		Filter(basicAuthFilter).
		Doc("Update a item.").
		To(UpdateItem).
		Reads(Item{}).
		Do(returnsInternalServerError, returnsNotFound, returnsUpdateSuccessful, returnsBadRequest))

	service.Route(service.POST("").
		Filter(basicAuthFilter).
		Doc("Insert a item into the database").
		To(CreateItem).
		Reads(Item{}).
		Returns(http.StatusOK, "Insert successful", "/item/{id}").
		Do(returnsInternalServerError, returnsBadRequest))

	service.Route(service.DELETE("/{id}").
		Filter(basicAuthFilter).
		Param(restful.PathParameter("id", "Item ID")).
		Doc("Delete a item").
		To(DeleteItem).
		Do(returnsInternalServerError, returnsNotFound, returnsDeleteSuccessful, returnsBadRequest))
	return service
}

func GetItemById(request *restful.Request, response *restful.Response) {
	sid := request.PathParameter("id")
	log.WithFields(log.Fields{"Requested ID": sid, "Path": request.SelectedRoutePath()}).Debug("Got Request")
	id, err := strconv.ParseUint(sid, 10, 64)
	if err != nil {
		response.WriteErrorString(http.StatusBadRequest, ERROR_INVALID_ID)
		log.WithFields(log.Fields{"Error Msg": err}).
			Info(ERROR_INVALID_ID)
		return
	}
	itm, err := getItemById(id)
	if err != nil {
		response.WriteErrorString(http.StatusNotFound, ERROR_INVALID_ID)
		log.WithFields(log.Fields{"Error Msg": err}).
			Info(ERROR_INVALID_ID)
		return
	}
	response.WriteEntity(itm)
}

func getItemById(id uint64) (Item, error) {
	res := Item{}
	err := iCol.Find(bson.M{"eid": id}).One(&res)
	return res, err
}

func GetItemLog(request *restful.Request, response *restful.Response) {
	sid := request.PathParameter("id")
	id, err := strconv.ParseUint(sid, 10, 64)
	if err != nil {
		response.WriteErrorString(http.StatusNotFound, ERROR_INVALID_ID)
		log.Info(err)
		return
	}
	history, err := getItemLog(id)
	if err != nil {
		response.WriteErrorString(http.StatusNotFound, ERROR_INVALID_ID)
		log.Info(err)
		return
	}
	response.WriteEntity(history)
}

func getItemLog(id uint64) ([]ItemHistory, error) {
	res := make([]ItemHistory, 0)
	err := ihCol.Find(bson.M{"item.eid": id}).All(&res)
	for i := 0; i != len(res); i++ {
		res[i].Timestamp = res[i].ID.Time()
	}
	return res, err
}

func CreateItem(request *restful.Request, response *restful.Response) {
	itm := new(Item)
	err := request.ReadEntity(itm)
	if err != nil {
		response.WriteErrorString(http.StatusBadRequest, ERROR_INVALID_INPUT)
		log.WithFields(log.Fields{"Error Msg": err}).Info(ERROR_INVALID_INPUT)
		return
	}

	id, err := createItem(itm)
	if err != nil {
		response.WriteErrorString(http.StatusInternalServerError, ERROR_INTERNAL)
		log.WithFields(log.Fields{"Error Msg": err}).Warn(ERROR_INTERNAL)
		return
	}
	response.WriteEntity("/item/" + strconv.FormatUint(id, 10))
}

func createItem(itm *Item) (uint64, error) {
	itm.EID = idgen.GenerateID()
	log.WithFields(log.Fields{"ID": itm.EID}).Debug("Generated ID")
	err := iCol.Insert(itm)
	return itm.EID, err
}

func ListItem(request *restful.Request, response *restful.Response) {
	itm, err := listItem()
	if err != nil {
		response.WriteErrorString(http.StatusInternalServerError, ERROR_INTERNAL)
		log.WithFields(log.Fields{"Error Msg": err}).Warn(ERROR_INTERNAL)
		return
	}
	response.WriteEntity(itm)
}

func listItem() ([]Item, error) {
	itm := make([]Item, 0)
	err := iCol.Find(nil).All(&itm)
	return itm, err
}

func UpdateItem(request *restful.Request, response *restful.Response) {
	itm := new(Item)
	err := request.ReadEntity(itm)
	if err != nil {
		response.WriteErrorString(http.StatusInternalServerError, ERROR_INTERNAL)
		log.WithFields(log.Fields{"Error Msg": err}).Warn(ERROR_INTERNAL)
		return
	}
	i, err := getItemById(itm.EID)
	if err != nil {
		response.WriteErrorString(http.StatusInternalServerError, ERROR_INTERNAL)
		log.Warn(err)
		return
	}
	h := i.NewItemHistory(itm, request.Attribute("User").(string))

	err = updateItem(itm, h)
	if err != nil {
		response.WriteErrorString(http.StatusInternalServerError, ERROR_INTERNAL)
		log.Warn(err)
		return
	}
	response.WriteEntity(true)
	return
}

func updateItem(itm *Item, ih *ItemHistory) error {
	err := ihCol.Insert(ih)
	if err != nil {
		return err
	}
	return iCol.Update(bson.M{"eid": itm.EID}, itm)
}

func checkItemExistance(itm *Item) bool {
	temp := Item{}
	err := iCol.Find(bson.M{"eid": itm.EID}).One(&temp)
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

func DeleteItem(request *restful.Request, response *restful.Response) {
	sid := request.PathParameter("id")
	id, err := strconv.ParseUint(sid, 10, 64)
	if err != nil {
		//TODO
		log.WithFields(log.Fields{"Error Msg": err}).Info(ERROR_INVALID_ID)
		response.WriteErrorString(http.StatusNotFound, ERROR_INVALID_ID)
		return
	}

	i, err := getItemById(id)
	if err != nil {
		if err.Error() == "not found" {
			log.WithFields(log.Fields{"ID": id}).Info(ERROR_INVALID_ID)
			response.WriteErrorString(http.StatusNotFound, ERROR_INVALID_ID)
			return
		}
		log.WithFields(log.Fields{"Error Msg": err}).Info(ERROR_INTERNAL)
		response.WriteErrorString(http.StatusInternalServerError, ERROR_INTERNAL)
		return
	}

	h := i.NewItemHistory(nil, request.Attribute("User").(string))
	err = deleteItem(&i, h)
	if err != nil {
		log.Warn(err)
		response.WriteErrorString(http.StatusInternalServerError, ERROR_INTERNAL)
		return
	}
	response.WriteEntity(true)
}

func deleteItem(itm *Item, ih *ItemHistory) error {
	err := ihCol.Insert(ih)
	if err != nil {
		return err
	}
	return iCol.Remove(bson.M{"eid": itm.EID})
}
