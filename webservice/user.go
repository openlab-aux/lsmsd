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
	log "github.com/sirupsen/logrus"
	"github.com/emicklei/go-restful"
	//"html/template"
	"net/http"
	//"strconv"
	//"strings"
	db "github.com/openlab-aux/lsmsd/database"
)

type UserWebService struct {
	d *db.UserDBProvider
	S *restful.WebService
	a *BasicAuthService
}

func NewUserService(d *db.UserDBProvider, a *BasicAuthService) *UserWebService {
	res := new(UserWebService)
	res.d = d
	res.a = a
	service := new(restful.WebService)
	service.
		Path("/users").
		Doc("User related services").
		ApiVersion("0.1").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	service.Route(service.GET("/{name}").
		Param(restful.PathParameter("name", "User identifier")).
		Doc("Returns public accessable information about the referenced user").
		To(res.GetUserByName).
		Writes(db.User{}).
		Do(returnsInternalServerError, returnsNotFound, returnsBadRequest))

	service.Route(service.GET("/{name}/log").
		Param(restful.PathParameter("name", "User identifier")).
		Doc("Returns all changes the user has made in the DB").
		To(res.GetUserLogByName).
		Writes(db.UserActionHistory{}).
		Do(returnsInternalServerError, returnsNotFound, returnsBadRequest))

	service.Route(service.GET("").
		Doc("List all users (this may be replaced by a paginated version)").
		To(res.ListUser).
		Writes([]db.User{}).
		Do(returnsInternalServerError))

	service.Route(service.PUT("").
		Filter(res.a.Auth).
		Doc("Update user information").
		To(res.UpdateUser).
		Reads(db.User{}).
		Do(returnsInternalServerError, returnsNotFound, returnsUpdateSuccessful, returnsBadRequest))

	service.Route(service.POST("").
		Doc("Register a new user").
		To(res.CreateUser).
		Reads(db.User{}).
		Returns(http.StatusOK, "Insert successful", "/users/{name}").
		Returns(http.StatusUnauthorized, "This username is not available", nil).
		Do(returnsInternalServerError, returnsBadRequest))

	service.Route(service.DELETE("/{name}").
		Filter(res.a.Auth).
		Param(restful.PathParameter("name", "User identifier")).
		Doc("Delete a user").
		To(res.DeleteUser).
		Returns(http.StatusForbidden, "Request not allowed; You may only delete your own account", nil).
		Do(returnsInternalServerError, returnsNotFound, returnsDeleteSuccessful, returnsBadRequest))
	res.S = service
	return res
}

func (p *UserWebService) GetUserByName(request *restful.Request, response *restful.Response) {
	name := request.PathParameter("name")
	user, err := p.d.GetUserByName(name)
	if err != nil {
		response.WriteErrorString(http.StatusNotFound, ERROR_INVALID_ID)
		log.WithFields(log.Fields{"Error Msg": err}).
			Info(ERROR_INVALID_ID)
		return
	}
	response.WriteEntity(user)
}

func (p *UserWebService) GetUserLogByName(request *restful.Request, response *restful.Response) {
	name := request.PathParameter("name")
	ul, err := p.d.GetUserLogByName(name)
	if err != nil {
		response.WriteErrorString(http.StatusNotFound, ERROR_INVALID_ID)
		log.Info(err)
		return
	}
	response.WriteEntity(ul)
	return
}

func (p *UserWebService) ListUser(request *restful.Request, response *restful.Response) {
	usr, err := p.d.ListUser()
	if err != nil {
		response.WriteErrorString(http.StatusInternalServerError, ERROR_INTERNAL)
		log.WithFields(log.Fields{"Error Msg": err}).Warn(ERROR_INTERNAL)
		return
	}
	response.WriteEntity(usr)
}

func (p *UserWebService) UpdateUser(request *restful.Request, response *restful.Response) {
	usr := new(db.User)
	err := request.ReadEntity(usr)
	if err != nil {
		response.WriteErrorString(http.StatusBadRequest, ERROR_INVALID_INPUT)
		log.WithFields(log.Fields{"Error Msg": err}).Warn(ERROR_INVALID_INPUT)
		return
	}
	if usr.Name != request.Attribute("User").(string) {
		log.WithFields(log.Fields{"User": request.Attribute("User").(string), "attempted to update": usr.Name}).Warn("Unauthorized update request")
		response.WriteErrorString(http.StatusForbidden, "Request not allowed")
		return
	}

	ex := p.d.CheckUserExistance(usr)
	if ex {
		if usr.Password != "" {
			log.Debug("User supplied new password.")
			err = usr.Secret.SetPassword(usr.Password)
		} else {
			temp, err := p.d.GetUserByName(usr.Name)
			if err != nil {
			} else {
				usr.Secret = temp.Secret // if no new password will be set, preserve old
			}
		}
		if err != nil { //fall through to error handling
		} else {
			err = p.d.UpdateUser(usr)
		}
		if err != nil {
			response.WriteErrorString(http.StatusInternalServerError, ERROR_INTERNAL)
			log.WithFields(log.Fields{"Err": err}).Warn("Error while updating User")
			return
		}
		response.WriteEntity(true)
		return
	} else {
		response.WriteErrorString(http.StatusNotFound, "User not found. Please register first")
		return
	}
}

func (p *UserWebService) CreateUser(request *restful.Request, response *restful.Response) {
	usr := new(db.User)
	err := request.ReadEntity(usr)
	log.WithFields(log.Fields{"Username": usr.Name}).Info("Attempted user registration")
	log.Debug(usr)
	if err != nil || usr.Password == "" {
		response.WriteErrorString(http.StatusBadRequest, ERROR_INVALID_INPUT)
		log.WithFields(log.Fields{"Error Msg": err}).Info(ERROR_INVALID_INPUT)
		return
	}

	ex := p.d.CheckUserExistance(usr)

	if ex {
		response.WriteErrorString(http.StatusForbidden, "This username is not available")
		return
	}
	err = usr.Secret.SetPassword(usr.Password)
	if err != nil {
	} else {

		err = p.d.CreateUser(usr)
	}
	if err != nil {
		response.WriteErrorString(http.StatusInternalServerError, ERROR_INTERNAL)
		log.WithFields(log.Fields{"Error Msg": err}).Warn(ERROR_INTERNAL)
		return
	}
	response.WriteEntity("/users/" + usr.Name)
}

func (p *UserWebService) DeleteUser(request *restful.Request, response *restful.Response) {
	name := request.PathParameter("name")
	log.WithFields(log.Fields{"Name": name}).Info("Got user DELETE request")

	if name != request.Attribute("User").(string) {
		log.WithFields(log.Fields{"User": request.Attribute("User").(string), "attempted to delete": name}).Warn("Unauthorized deletion request")
		response.WriteErrorString(http.StatusForbidden, "Request not allowed")
		return
	}

	err := p.d.DeleteUser(name)
	if err != nil {
		log.WithFields(log.Fields{"Error Msg": err}).Info(ERROR_INTERNAL)
		response.WriteErrorString(http.StatusInternalServerError, ERROR_INTERNAL)
		return
	}
	response.WriteEntity(true)
}
