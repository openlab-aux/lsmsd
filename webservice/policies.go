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
	log "github.com/Sirupsen/logrus"
	"github.com/emicklei/go-restful"
	//"html/template"
	"net/http"
	//"strconv"
	//"strings"
	db "github.com/openlab-aux/lsmsd/database"
)

type PolicyWebService struct {
	d *db.PolicyDBProvider
	S *restful.WebService
	a *BasicAuthService
}

func NewPolicyService(d *db.PolicyDBProvider, a *BasicAuthService) *PolicyWebService {
	res := new(PolicyWebService)
	res.d = d
	res.a = a

	service := new(restful.WebService)
	service.
		Path("/policy").
		Doc("Policy related services").
		ApiVersion("0.1").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	service.Route(service.GET("/{name}").
		Param(restful.PathParameter("name", "Policy Name")).
		Doc("Returns a single policy identified by its name").
		To(res.GetPolicyByName).
		Writes(db.Policy{}).
		Do(returnsInternalServerError, returnsNotFound, returnsBadRequest))

	service.Route(service.GET("/{name}/log").
		Param(restful.PathParameter("name", "Policy Name")).
		Doc("Returns the policys changelog").
		To(res.GetPolicyLog).
		Writes(db.PolicyHistory{}).
		Do(returnsInternalServerError, returnsNotFound, returnsBadRequest))

	service.Route(service.GET("").
		Doc("List all available policys (this may be replaced by a paginated version)").
		To(res.ListPolicy).
		Writes([]db.Policy{}).
		Do(returnsInternalServerError))

	service.Route(service.PUT("").
		Filter(res.a.Auth).
		Doc("Update a policy").
		To(res.UpdatePolicy).
		Reads(db.Policy{}).
		Do(returnsInternalServerError, returnsNotFound, returnsUpdateSuccessful, returnsBadRequest))

	service.Route(service.POST("").
		Filter(res.a.Auth).
		Doc("Insert a policy").
		To(res.CreatePolicy).
		Reads(db.Policy{}).
		Returns(http.StatusOK, "Insert successful", "/policy/{name").
		Do(returnsInternalServerError, returnsBadRequest))

	service.Route(service.DELETE("/{name}").
		Filter(res.a.Auth).
		Param(restful.PathParameter("name", "Policy Name")).
		Doc("Delete a policy").
		To(res.DeletePolicy).
		Do(returnsInternalServerError, returnsNotFound, returnsDeleteSuccessful, returnsBadRequest))
	res.S = service
	return res
}

func (p *PolicyWebService) GetPolicyByName(request *restful.Request, response *restful.Response) {
	log.WithFields(log.Fields{"Path": request.SelectedRoutePath()}).Debug("Got Request")
	name := request.PathParameter("name")
	pol, err := p.d.GetPolicyByName(name)
	if err != nil {
		response.WriteErrorString(http.StatusNotFound, ERROR_INVALID_ID)
		log.WithFields(log.Fields{"Error Msg": err}).
			Info(ERROR_INVALID_ID)
		return
	}
	response.WriteEntity(pol)
}

func (p *PolicyWebService) GetPolicyLog(request *restful.Request, response *restful.Response) {
	name := request.PathParameter("name")

	history, err := p.d.GetPolicyLog(name)
	if err != nil {
		response.WriteErrorString(http.StatusNotFound, ERROR_INVALID_ID)
		log.Info(err)
		return
	}
	response.WriteEntity(history)
}

func (p *PolicyWebService) ListPolicy(request *restful.Request, response *restful.Response) {
	pol, err := p.d.ListPolicy()
	if err != nil {
		response.WriteErrorString(http.StatusInternalServerError, ERROR_INTERNAL)
		log.WithFields(log.Fields{"Error Msg": err}).Warn(ERROR_INTERNAL)
		return
	}
	response.WriteEntity(pol)
}

func (p *PolicyWebService) UpdatePolicy(request *restful.Request, response *restful.Response) {
	pol := new(db.Policy)
	err := request.ReadEntity(pol)
	if err != nil {
		response.WriteErrorString(http.StatusBadRequest, ERROR_INVALID_ID)
		log.WithFields(log.Fields{"Error Msg": err}).Warn(ERROR_INVALID_ID)
		return
	}
	po, err := p.d.GetPolicyByName(pol.Name)
	if err != nil {
		response.WriteErrorString(http.StatusInternalServerError, ERROR_INTERNAL)
		log.Warn(err)
		return
	}
	h := po.NewPolicyHistory(pol, request.Attribute("User").(string))

	err = p.d.UpdatePolicy(pol, h)
	if err != nil {
		response.WriteErrorString(http.StatusInternalServerError, ERROR_INTERNAL)
		log.Warn(err)
		return
	}
	response.WriteEntity(true)
}

func (p *PolicyWebService) CreatePolicy(request *restful.Request, response *restful.Response) {
	pol := new(db.Policy)
	err := request.ReadEntity(pol)
	if err != nil {
		response.WriteErrorString(http.StatusBadRequest, ERROR_INVALID_INPUT)
		log.WithFields(log.Fields{"Error Msg": err}).Info(ERROR_INVALID_INPUT)
		return
	}

	ex := p.d.CheckPolicyExistance(pol)

	if ex {
		response.WriteErrorString(http.StatusUnauthorized, "This Policy does already exist")
		return
	}

	err = p.d.CreatePolicy(pol)
	if err != nil {
		response.WriteErrorString(http.StatusInternalServerError, ERROR_INTERNAL)
		log.WithFields(log.Fields{"Error Msg": err}).Warn(ERROR_INTERNAL)
		return
	}

	response.WriteEntity("/policy/" + pol.Name)
}

func (p *PolicyWebService) DeletePolicy(request *restful.Request, response *restful.Response) {
	name := request.PathParameter("name")
	po, err := p.d.GetPolicyByName(name)
	if err != nil {
		if err.Error() == "not found" {
			log.Info(err)
			response.WriteErrorString(http.StatusNotFound, ERROR_INVALID_ID)
			return
		}
		log.Warn(err)
		response.WriteErrorString(http.StatusInternalServerError, ERROR_INTERNAL)
		return
	}

	h := po.NewPolicyHistory(nil, request.Attribute("User").(string))
	err = p.d.DeletePolicy(&po, h)
	if err != nil {
		log.WithFields(log.Fields{"Error Msg": err}).Info(ERROR_INTERNAL)
		response.WriteErrorString(http.StatusInternalServerError, ERROR_INTERNAL)
		return
	}
	response.WriteEntity(true)
}
