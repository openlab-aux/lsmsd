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
	"gopkg.in/mgo.v2/bson"
	"net/http"
	"strconv"
	//"strings"
)

type Item struct {
	ID          bson.ObjectId `bson:"_id,omitempty" json:"-"`
	EID         uint64
	Name        string          `bson:",omitempty"`
	Description string          `bson:",omitempty"`
	Contains    []bson.ObjectId `bson:",omitempty"`
	Owner       string          `bson:",omitempty"`
	Maintainer  string          `bson:",omitempty"`
	Usage       string          `bson:",omitempty"`
	Discard     string          `bson:",omitempty"`
}

type ItemHistory struct {
	ID   bson.ObjectId `bson:"_id,omitempty" json:"-"`
	Item Item
}

func NewItemService() *restful.WebService {
	service := new(restful.WebService)
	service.
		Path("/item").
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON, restful.MIME_XML)

	service.Route(service.GET("/{id}").To(GetItemById))
	service.Route(service.GET("/{id}/log").To(GetItemLog))
	service.Route(service.GET("").To(ListItem))
	service.Route(service.PUT("").Filter(basicAuthFilter).To(UpdateItem))
	service.Route(service.POST("").Filter(basicAuthFilter).To(CreateItem))
	service.Route(service.DELETE("/{id}").Filter(basicAuthFilter).To(DeleteItem))
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
		response.WriteErrorString(http.StatusBadRequest, ERROR_INVALID_ID)
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
	err := ihCol.Find(bson.M{"item": bson.M{"eid": id}}).All(&res)
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
	ex := checkItemExistance(itm)
	if ex {
		err = updateItem(itm)
		if err != nil {
			response.WriteErrorString(http.StatusInternalServerError, ERROR_INTERNAL)
			log.WithFields(log.Fields{"Err": err}).Warn("Error while updating Item")
			return
		}
		response.WriteEntity(true)
		return
	} else {
		response.WriteErrorString(http.StatusNotFound, "Item not found")
		return
	}
}

func updateItem(itm *Item) error {
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
		log.WithFields(log.Fields{"Error Msg": err}).Info(ERROR_INVALID_ID)
		response.WriteErrorString(http.StatusNotFound, ERROR_INVALID_ID)
		return
	}

	err = deleteItem(id)
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
	response.WriteEntity(true)
}

func deleteItem(id uint64) error {
	return iCol.Remove(bson.M{"eid": id})
}
