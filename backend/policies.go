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
	"html/template"
	"net/http"
	"strconv"
	//"strings"
)

type Policy struct {
	ID          uint64
	Name        string
	Description string
}

func NewPolicyService() *restful.WebService {
	service := new(restful.WebService)
	service.
		Path("/policy").
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON, restful.MIME_XML)

	service.Route(service.GET("/{id}").To(GetPolicyById))
	service.Route(service.GET("").To(ListPolicy))
	service.Route(service.PUT("").To(UpdatePolicy))
	service.Route(service.PUT("/{id}").To(UpdatePolicy))
	service.Route(service.POST("").To(CreatePolicy))
	service.Route(service.DELETE("/{id}").To(DeletePolicy))
	return service
}

func GetPolicyById(request *restful.Request, response *restful.Response) {
	sid := request.PathParameter("id")
	id, err := strconv.ParseUint(sid, 10, 64)
	if err != nil {
		response.WriteErrorString(http.StatusBadRequest, ERROR_INVALID_ID)
		log.WithFields(log.Fields{"Error Msg": err}).Info(ERROR_INVALID_ID)
		return
	}
	pol, err := getPolicyById(id)
	if err != nil || pol.ID == 0 {
		response.WriteErrorString(http.StatusBadRequest, ERROR_INVALID_ID)
		log.WithFields(log.Fields{"Error Msg": err}).
			Info(ERROR_INVALID_ID)
		return
	}
	response.WriteEntity(pol)
}

func getPolicyById(id uint64) (Policy, error) {
	stmt, err := db.Prepare("SELECT * FROM policies WHERE id=?;")
	if err != nil {
		log.WithFields(log.Fields{"Error Msg": err}).
			Fatal(ERROR_STMT_PREPARE)
	}
	defer stmt.Close()
	row := stmt.QueryRow(id)
	pol := new(Policy)
	err = row.Scan(&pol.ID, &pol.Name, &pol.Description)
	if err != nil {
		log.WithFields(log.Fields{"Error Msg": err}).
			Warn("Scan Error")
	}
	return *pol, nil
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
	stmt, err := db.Prepare("SELECT * FROM policies;")
	if err != nil {
		log.WithFields(log.Fields{"Error Msg": err}).Fatal(ERROR_STMT_PREPARE)
		return nil, err
	}
	defer stmt.Close()
	rows, err := stmt.Query()
	if err != nil {
		log.WithFields(log.Fields{"Error Msg": err}).Warn(ERROR_INTERNAL)
		return nil, err
	}
	result := make([]Policy, 0)
	for rows.Next() {
		temp := new(Policy)
		rows.Scan(&temp.ID, &temp.Name, &temp.Description)
		result = append(result, *temp)
	}

	return result, nil
}

func UpdatePolicy(request *restful.Request, response *restful.Response) {
	pol := new(Policy)
	err := request.ReadEntity(pol)
	if err != nil {
		response.WriteErrorString(http.StatusInternalServerError, ERROR_INTERNAL)
		log.WithFields(log.Fields{"Error Msg": err}).Warn(ERROR_INTERNAL)
		return
	}
	stmt, err := db.Prepare(`
		SELECT * FROM policies ORDER BY id DESC LIMIT 0, 1;`)
	if err != nil {
		response.WriteErrorString(http.StatusInternalServerError, ERROR_INTERNAL)
		log.WithFields(log.Fields{"Error Msg": err}).Warn(ERROR_INTERNAL)
		return
	}
	defer stmt.Close()
	row := stmt.QueryRow()
	var id uint64
	temp := new(Policy)
	err = row.Scan(&id, &temp.Name, &temp.Description)
	if err != nil {
		response.WriteErrorString(http.StatusInternalServerError, ERROR_INTERNAL)
		log.WithFields(log.Fields{"Error Msg": err}).Warn(ERROR_INTERNAL)
		return
	}

	log.WithFields(log.Fields{"Id": id}).Debug("Highest ID")

	if pol.ID <= id {
		err := updatePolicy(pol)
		if err != nil {
			response.WriteErrorString(http.StatusInternalServerError, ERROR_INTERNAL)
			return
		} else {
			response.WriteEntity(true)
		}
	} else {
		CreatePolicy(request, response)
	}
}

func updatePolicy(pol *Policy) error {
	stmt, err := db.Prepare(`UPDATE policies SET name=?, description=? WHERE id=?;`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(template.HTMLEscapeString(pol.Name), template.HTMLEscapeString(pol.Description), pol.ID)
	if err != nil {
		return err
	}
	return nil
}

func CreatePolicy(request *restful.Request, response *restful.Response) {
	pol := new(Policy)
	err := request.ReadEntity(pol)
	if err != nil {
		response.WriteErrorString(http.StatusBadRequest, ERROR_INVALID_INPUT)
		log.WithFields(log.Fields{"Error Msg": err}).Info(ERROR_INVALID_INPUT)
		return
	}

	id, err := createPolicy(pol)
	if err != nil {
		response.WriteErrorString(http.StatusInternalServerError, ERROR_INTERNAL)
		log.WithFields(log.Fields{"Error Msg": err}).Warn(ERROR_INTERNAL)
		return
	}
	response.WriteEntity("/policy/" + strconv.FormatUint(id, 10))
}

func createPolicy(pol *Policy) (uint64, error) {
	stmt, err := db.Prepare(`
		INSERT INTO policies ('name', 'description') VALUES (?, ?);`)
	if err != nil {
		log.WithFields(log.Fields{"Error Msg": err}).Fatal(ERROR_STMT_PREPARE)
		return 0, err
	}
	defer stmt.Close()
	res, err := stmt.Exec(template.HTMLEscapeString(pol.Name), template.HTMLEscapeString(pol.Description))
	if err != nil {
		log.WithFields(log.Fields{"Error Msg": err}).Warn(ERROR_INSERT)
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		log.WithFields(log.Fields{"Error Msg": err}).Warn(ERROR_QUERY)
		return 0, err
	}
	return uint64(id), nil
}

func DeletePolicy(request *restful.Request, response *restful.Response) {
	sid := request.PathParameter("id")
	id, err := strconv.ParseUint(sid, 10, 64)
	if err != nil {
		log.WithFields(log.Fields{"Error Msg": err}).Info(ERROR_INVALID_ID)
		response.WriteErrorString(http.StatusBadRequest, ERROR_INVALID_ID)
		return
	}

	err = deletePolicy(id)
	if err != nil {
		log.WithFields(log.Fields{"Error Msg": err}).Info(ERROR_INTERNAL)
		response.WriteErrorString(http.StatusInternalServerError, ERROR_INTERNAL)
		return
	}
	response.WriteEntity(true)
}

func deletePolicy(id uint64) error {
	stmt, err := db.Prepare("DELETE FROM policies WHERE id=?;")
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(id)
	if err != nil {
		return err
	}
	return nil
}
