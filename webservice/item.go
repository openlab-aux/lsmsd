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

package webservice

import (
	//	"database/sql"
	log "github.com/Sirupsen/logrus"
	"github.com/emicklei/go-restful"
	//"html/template"
	//	"github.com/fatih/structs"
	"net/http"
	"strconv"
	//"strings"
	db "github.com/openlab-aux/lsmsd/database"
)

type ItemWebService struct {
	d *db.ItemDBProvider
	S *restful.WebService
	a *BasicAuthService
}

func NewItemWebService(d *db.ItemDBProvider, a *BasicAuthService) *ItemWebService {
	res := new(ItemWebService)
	res.d = d
	res.a = a

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
		To(res.GetItemById).
		Writes(db.Item{}).
		Do(returnsInternalServerError, returnsNotFound, returnsBadRequest))

	service.Route(service.GET("/{id}/log").
		Param(restful.PathParameter("id", "Item ID")).
		Doc("Returns the items changelog").
		To(res.GetItemLog).
		Writes(db.ItemHistory{}).
		Do(returnsInternalServerError, returnsNotFound, returnsBadRequest))

	service.Route(service.GET("").
		Doc("List all available items (this may be replaced by a paginated version)").
		To(res.ListItem).
		Writes([]db.Item{}).
		Do(returnsInternalServerError))

	service.Route(service.PUT("").
		Filter(res.a.Auth).
		Doc("Update a item.").
		To(res.UpdateItem).
		Reads(db.Item{}).
		Do(returnsInternalServerError, returnsNotFound, returnsUpdateSuccessful, returnsBadRequest))

	service.Route(service.POST("").
		Filter(res.a.Auth).
		Doc("Insert a item into the database").
		To(res.CreateItem).
		Reads(db.Item{}).
		Returns(http.StatusOK, "Insert successful", "/item/{id}").
		Do(returnsInternalServerError, returnsBadRequest))

	service.Route(service.DELETE("/{id}").
		Filter(res.a.Auth).
		Param(restful.PathParameter("id", "Item ID")).
		Doc("Delete a item").
		To(res.DeleteItem).
		Do(returnsInternalServerError, returnsNotFound, returnsDeleteSuccessful, returnsBadRequest))

	res.S = service
	return res
}

func (s *ItemWebService) GetItemById(request *restful.Request, response *restful.Response) {
	sid := request.PathParameter("id")
	log.WithFields(log.Fields{"Requested ID": sid, "Path": request.SelectedRoutePath()}).Debug("Got Request")
	id, err := strconv.ParseUint(sid, 10, 64)
	if err != nil {
		response.WriteErrorString(http.StatusBadRequest, ERROR_INVALID_ID)
		log.WithFields(log.Fields{"Error Msg": err}).
			Info(ERROR_INVALID_ID)
		return
	}
	itm, err := s.d.GetItemById(id)
	if err != nil {
		response.WriteErrorString(http.StatusNotFound, ERROR_INVALID_ID)
		log.WithFields(log.Fields{"Error Msg": err}).
			Info(ERROR_INVALID_ID)
		return
	}
	response.WriteEntity(itm)
}

func (s *ItemWebService) GetItemLog(request *restful.Request, response *restful.Response) {
	sid := request.PathParameter("id")
	id, err := strconv.ParseUint(sid, 10, 64)
	if err != nil {
		response.WriteErrorString(http.StatusNotFound, ERROR_INVALID_ID)
		log.Info(err)
		return
	}
	history, err := s.d.GetItemLog(id)
	if err != nil {
		response.WriteErrorString(http.StatusNotFound, ERROR_INVALID_ID)
		log.Info(err)
		return
	}
	response.WriteEntity(history)
}

func (s *ItemWebService) CreateItem(request *restful.Request, response *restful.Response) {
	itm := new(db.Item)
	err := request.ReadEntity(itm)
	if err != nil {
		response.WriteErrorString(http.StatusBadRequest, ERROR_INVALID_INPUT)
		log.WithFields(log.Fields{"Error Msg": err}).Info(ERROR_INVALID_INPUT)
		return
	}

	id, err := s.d.CreateItem(itm)
	if err != nil {
		response.WriteErrorString(http.StatusInternalServerError, ERROR_INTERNAL)
		log.WithFields(log.Fields{"Error Msg": err}).Warn(ERROR_INTERNAL)
		return
	}
	response.WriteEntity("/item/" + strconv.FormatUint(id, 10))
}

func (s *ItemWebService) ListItem(request *restful.Request, response *restful.Response) {
	itm, err := s.d.ListItem()
	if err != nil {
		response.WriteErrorString(http.StatusInternalServerError, ERROR_INTERNAL)
		log.WithFields(log.Fields{"Error Msg": err}).Warn(ERROR_INTERNAL)
		return
	}
	response.WriteEntity(itm)
}

func (s *ItemWebService) UpdateItem(request *restful.Request, response *restful.Response) {
	itm := new(db.Item)
	err := request.ReadEntity(itm)
	if err != nil {
		response.WriteErrorString(http.StatusInternalServerError, ERROR_INTERNAL)
		log.WithFields(log.Fields{"Error Msg": err}).Warn(ERROR_INTERNAL)
		return
	}
	i, err := s.d.GetItemById(itm.EID)
	if err != nil {
		response.WriteErrorString(http.StatusInternalServerError, ERROR_INTERNAL)
		log.Warn(err)
		return
	}
	h := i.NewItemHistory(itm, request.Attribute("User").(string))

	err = s.d.UpdateItem(itm, h)
	if err != nil {
		response.WriteErrorString(http.StatusInternalServerError, ERROR_INTERNAL)
		log.Warn(err)
		return
	}
	response.WriteEntity(true)
	return
}

func (s *ItemWebService) DeleteItem(request *restful.Request, response *restful.Response) {
	sid := request.PathParameter("id")
	id, err := strconv.ParseUint(sid, 10, 64)
	if err != nil {
		//TODO
		log.WithFields(log.Fields{"Error Msg": err}).Info(ERROR_INVALID_ID)
		response.WriteErrorString(http.StatusNotFound, ERROR_INVALID_ID)
		return
	}

	i, err := s.d.GetItemById(id)
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
	err = s.d.DeleteItem(&i, h)
	if err != nil {
		log.Warn(err)
		response.WriteErrorString(http.StatusInternalServerError, ERROR_INTERNAL)
		return
	}
	response.WriteEntity(true)
}
