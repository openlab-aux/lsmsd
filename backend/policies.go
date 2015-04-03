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
	"github.com/emicklei/go-restful"
	//"html/template"
	"net/http"
	//"strconv"
	//"strings"
	dmp "github.com/sergi/go-diff/diffmatchpatch"
	"gopkg.in/mgo.v2/bson"
	"time"
)

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

func NewPolicyService() *restful.WebService {
	service := new(restful.WebService)
	service.
		Path("/policy").
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON, restful.MIME_XML)

	service.Route(service.GET("/{name}").To(GetPolicyByName))
	service.Route(service.GET("/{name}/log").To(GetPolicyLog))
	service.Route(service.GET("").To(ListPolicy))
	service.Route(service.PUT("").Filter(basicAuthFilter).To(UpdatePolicy))
	service.Route(service.POST("").Filter(basicAuthFilter).To(CreatePolicy))
	service.Route(service.DELETE("/{name}").Filter(basicAuthFilter).To(DeletePolicy))
	return service
}

func GetPolicyByName(request *restful.Request, response *restful.Response) {
	log.WithFields(log.Fields{"Path": request.SelectedRoutePath()}).Debug("Got Request")
	name := request.PathParameter("name")
	pol, err := getPolicyByName(name)
	if err != nil {
		response.WriteErrorString(http.StatusNotFound, ERROR_INVALID_ID)
		log.WithFields(log.Fields{"Error Msg": err}).
			Info(ERROR_INVALID_ID)
		return
	}
	response.WriteEntity(pol)
}

func getPolicyByName(name string) (Policy, error) {
	res := Policy{}
	err := pCol.Find(bson.M{"name": name}).One(&res)
	return res, err
}

func GetPolicyLog(request *restful.Request, response *restful.Response) {
	name := request.PathParameter("name")
	history, err := getPolicyLog(name)
	if err != nil {
		response.WriteErrorString(http.StatusNotFound, ERROR_INVALID_ID)
		log.Info(err)
		return
	}
	response.WriteEntity(history)
}

func getPolicyLog(name string) ([]PolicyHistory, error) {
	res := make([]PolicyHistory, 0)
	err := phCol.Find(bson.M{"policy.name": name}).All(&res)
	for i := 0; i != len(res); i++ {
		res[i].Timestamp = res[i].ID.Time()
	}
	return res, err
}

func ListPolicy(request *restful.Request, response *restful.Response) {
	pol, err := listPolicy()
	if err != nil {
		response.WriteErrorString(http.StatusInternalServerError, ERROR_INTERNAL)
		log.WithFields(log.Fields{"Error Msg": err}).Warn(ERROR_INTERNAL)
		return
	}
	response.WriteEntity(pol)
}

func listPolicy() ([]Policy, error) {
	pol := make([]Policy, 0)
	err := pCol.Find(nil).All(&pol)
	return pol, err
}

func UpdatePolicy(request *restful.Request, response *restful.Response) {
	pol := new(Policy)
	err := request.ReadEntity(pol)
	if err != nil {
		response.WriteErrorString(http.StatusInternalServerError, ERROR_INTERNAL)
		log.WithFields(log.Fields{"Error Msg": err}).Warn(ERROR_INTERNAL)
		return
	}
	p, err := getPolicyByName(pol.Name)
	if err != nil {
		response.WriteErrorString(http.StatusInternalServerError, ERROR_INTERNAL)
		log.Warn(err)
		return
	}
	h := p.NewPolicyHistory(pol, request.Attribute("User").(string))

	err = updatePolicy(pol, h)
	if err != nil {
		response.WriteErrorString(http.StatusInternalServerError, ERROR_INTERNAL)
		log.Warn(err)
		return
	}
	response.WriteEntity(true)
}

func updatePolicy(pol *Policy, ph *PolicyHistory) error {
	err := phCol.Insert(ph)
	if err != nil {
		return err
	}
	return pCol.Update(bson.M{"name": pol.Name}, pol)
}

func checkPolicyExistance(pol *Policy) bool {
	temp := Policy{}
	err := pCol.Find(bson.M{"name": pol.Name}).One(&temp)
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

func CreatePolicy(request *restful.Request, response *restful.Response) {
	pol := new(Policy)
	err := request.ReadEntity(pol)
	if err != nil {
		response.WriteErrorString(http.StatusBadRequest, ERROR_INVALID_INPUT)
		log.WithFields(log.Fields{"Error Msg": err}).Info(ERROR_INVALID_INPUT)
		return
	}

	ex := checkPolicyExistance(pol)

	if ex {
		response.WriteErrorString(http.StatusUnauthorized, "This Policy does already exist")
		return
	}

	err = createPolicy(pol)
	if err != nil {
		response.WriteErrorString(http.StatusInternalServerError, ERROR_INTERNAL)
		log.WithFields(log.Fields{"Error Msg": err}).Warn(ERROR_INTERNAL)
		return
	}

	response.WriteEntity("/policy/" + pol.Name)
}

func createPolicy(pol *Policy) error {
	return pCol.Insert(pol)
}

func DeletePolicy(request *restful.Request, response *restful.Response) {
	name := request.PathParameter("name")
	p, err := getPolicyByName(name)
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

	h := p.NewPolicyHistory(nil, request.Attribute("User").(string))
	err = deletePolicy(&p, h)
	if err != nil {
		log.WithFields(log.Fields{"Error Msg": err}).Info(ERROR_INTERNAL)
		response.WriteErrorString(http.StatusInternalServerError, ERROR_INTERNAL)
		return
	}
	response.WriteEntity(true)
}

func deletePolicy(p *Policy, ph *PolicyHistory) error {
	err := phCol.Insert(ph)
	if err != nil {
		return err
	}
	return pCol.Remove(bson.M{"name": p.Name})
}
