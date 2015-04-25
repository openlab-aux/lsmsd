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
	"bytes"
	"encoding/hex"
	db "github.com/openlab-aux/lsmsd/database"
	"gopkg.in/mgo.v2/bson"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io/ioutil"
)

type ItemWebService struct {
	d *db.ItemDBProvider
	S *restful.WebService
	a *BasicAuthService
	i *db.ImageDBProvider
	u *UpdateService
}

func NewItemWebService(d *db.ItemDBProvider, i *db.ImageDBProvider, a *BasicAuthService, u *UpdateService) *ItemWebService {
	res := new(ItemWebService)
	res.d = d
	res.a = a
	res.i = i
	res.u = u

	service := new(restful.WebService)
	service.
		Path("/items").
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

	service.Route(service.GET("/coffee").
		Doc("Obviously not an easteregg. Go away. Leave me alone.").
		To(res.NotAnEasterEgg).
		Returns(http.StatusTeapot, "There is no coffee", nil))

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
		Returns(http.StatusOK, "Insert successful", "/items/{id}").
		Do(returnsInternalServerError, returnsBadRequest))

	service.Route(service.POST("/{id}/image").
		Filter(res.a.Auth).
		Param(restful.BodyParameter("image", "Your png, jpeg or gif.")).
		Doc("Attach a image to this item").
		To(res.AttachImage).
		Consumes("image/png", "image/jpeg", "image/gif").
		Do(returnsInternalServerError, returnsNotFound, returnsBadRequest))

	service.Route(service.DELETE("/{id}/image/{imageid}").
		Filter(res.a.Auth).
		Param(restful.PathParameter("id", "Item ID")).
		Param(restful.PathParameter("imageid", "Image identifier")).
		Doc("Remove a image from this item").
		To(res.RemoveImage).
		Do(returnsInternalServerError, returnsNotFound, returnsBadRequest))

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
		response.WriteErrorString(http.StatusBadRequest, ERROR_INVALID_ID)
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
	response.WriteEntity("/items/" + strconv.FormatUint(id, 10))
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
		response.WriteErrorString(http.StatusBadRequest, ERROR_INVALID_ID)
		log.WithFields(log.Fields{"Error Msg": err}).Warn(ERROR_INVALID_ID)
		return
	}
	i, err := s.d.GetItemById(itm.EID)
	if err != nil {
		response.WriteErrorString(http.StatusNotFound, ERROR_INVALID_ID)
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

	s.u.PushUpdate(h)
	response.WriteEntity(true)
	return
}

func (s *ItemWebService) DeleteItem(request *restful.Request, response *restful.Response) {
	sid := request.PathParameter("id")
	id, err := strconv.ParseUint(sid, 10, 64)
	if err != nil {
		//TODO
		log.WithFields(log.Fields{"Error Msg": err}).Info(ERROR_INVALID_ID)
		response.WriteErrorString(http.StatusBadRequest, ERROR_INVALID_ID)
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
	s.u.PushUpdate(h)

	response.WriteEntity(true)
}

func (s *ItemWebService) NotAnEasterEgg(req *restful.Request, res *restful.Response) {
	res.WriteErrorString(http.StatusTeapot, "Try some mate tea")
	return
}

func (s *ItemWebService) AttachImage(req *restful.Request, res *restful.Response) {
	sid := req.PathParameter("id")
	id, err := strconv.ParseUint(sid, 10, 64)
	if err != nil {
		log.Debug(err)
		res.WriteErrorString(http.StatusBadRequest, ERROR_INVALID_ID)
		return
	}

	if !s.d.CheckItemExistance(&db.Item{EID: id}) {
		log.Debug(err)
		res.WriteErrorString(http.StatusNotFound, ERROR_INVALID_ID)
		return
	}

	_body, err := ioutil.ReadAll(req.Request.Body)

	if err != nil {
		log.Debug(err)
		res.WriteErrorString(http.StatusBadRequest, "")
		return
	}

	body := bytes.NewReader(_body)

	switch req.HeaderParameter("Content-Type") {
	case "image/png":
		log.Println("png")
		_, err := png.Decode(body)
		if err != nil {
			log.Debug(err)
			res.WriteErrorString(http.StatusBadRequest, "")
			return
		}
		body.Seek(0, 0)

	case "image/jpeg":
		log.Println("jpeg")
		_, err := jpeg.Decode(body)
		if err != nil {
			log.Debug(err)
			res.WriteErrorString(http.StatusBadRequest, "")
			return
		}
		body.Seek(0, 0)

	case "image/gif":
		log.Println("gif")
		_, err := gif.Decode(body)
		if err != nil {
			log.Debug(err)
			res.WriteErrorString(http.StatusBadRequest, "")
			return
		}
		body.Seek(0, 0)

	default:
		//TODO
		log.Debug("none")
	}
	im, err := s.i.Create(body, req.Attribute("User").(string), req.HeaderParameter("Content-Type"), id)
	if err != nil {
		log.Debug(err)
		res.WriteErrorString(http.StatusInternalServerError, ERROR_INTERNAL)
		return
	}
	err = s.d.AddImage(id, im, req.Attribute("User").(string))
	if err != nil {
		log.Debug(err)
		res.WriteErrorString(http.StatusInternalServerError, ERROR_INTERNAL)
		return
	}
	return
}

func (s *ItemWebService) RemoveImage(req *restful.Request, res *restful.Response) {
	sid := req.PathParameter("id")
	id, err := strconv.ParseUint(sid, 10, 64)
	if err != nil {
		log.Debug(err)
		res.WriteErrorString(http.StatusBadRequest, ERROR_INVALID_ID)
		return
	}

	simgid := req.PathParameter("imageid")
	imgid, err := hex.DecodeString(simgid)
	if err != nil {
		log.Debug(err)
		res.WriteErrorString(http.StatusBadRequest, ERROR_INVALID_ID)
		return
	}

	if !s.d.CheckItemExistance(&db.Item{EID: id}) {
		log.Debug(err)
		res.WriteErrorString(http.StatusNotFound, ERROR_INVALID_ID)
		return
	}

	err = s.i.Remove(bson.ObjectId(imgid))
	if err != nil {
		log.Debug(err)
		res.WriteErrorString(http.StatusInternalServerError, ERROR_INTERNAL)
		return
	}

	err = s.d.RemoveImage(id, bson.ObjectId(imgid), req.Attribute("User").(string))
	if err != nil {
		log.Debug(err)
		res.WriteErrorString(http.StatusInternalServerError, ERROR_INTERNAL)
		return
	}
}
