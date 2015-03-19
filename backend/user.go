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

type User struct {
	ID   uint64
	Name string
}

func NewUserService() *restful.WebService {
	service := new(restful.WebService)
	service.
		Path("/user").
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON, restful.MIME_XML)

	service.Route(service.GET("/{id}").To(GetUserById))
	service.Route(service.GET("").To(ListUser))
	service.Route(service.PUT("").To(UpdateUser))
	service.Route(service.PUT("/{id}").To(UpdateUser))
	service.Route(service.POST("").To(CreateUser))
	service.Route(service.DELETE("/{id}").To(DeleteUser))
	return service
}

func GetUserById(request *restful.Request, response *restful.Response) {
	sid := request.PathParameter("id")
	id, err := strconv.ParseUint(sid, 10, 64)
	if err != nil {
		response.WriteErrorString(http.StatusBadRequest, ERROR_INVALID_ID)
		log.WithFields(log.Fields{"Error Msg": err}).Info(ERROR_INVALID_ID)
		return
	}
	user, err := getUserById(id)
	if err != nil || user.ID == 0 {
		response.WriteErrorString(http.StatusBadRequest, ERROR_INVALID_ID)
		log.WithFields(log.Fields{"Error Msg": err}).
			Info(ERROR_INVALID_ID)
		return
	}
	response.WriteEntity(user)
}

func getUserById(id uint64) (User, error) {
	stmt, err := db.Prepare("SELECT * FROM user WHERE id=?;")
	if err != nil {
		log.WithFields(log.Fields{"Error Msg": err}).
			Fatal(ERROR_STMT_PREPARE)
	}
	defer stmt.Close()
	row := stmt.QueryRow(id)
	user := new(User)
	err = row.Scan(&user.ID, &user.Name)
	if err != nil {
		log.WithFields(log.Fields{"Error Msg": err}).
			Warn("Scan Error")
	}
	return *user, nil
}

func ListUser(request *restful.Request, response *restful.Response) {
	usr, err := listUser()
	if err != nil {
		response.WriteErrorString(http.StatusInternalServerError, ERROR_INTERNAL)
		log.WithFields(log.Fields{"Error Msg": err}).Warn(ERROR_INTERNAL)
		return
	}
	response.WriteEntity(usr)
}

func listUser() ([]User, error) {
	stmt, err := db.Prepare("SELECT * FROM user;")
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
	result := make([]User, 0)
	for rows.Next() {
		temp := new(User)
		rows.Scan(&temp.ID, &temp.Name)
		result = append(result, *temp)
	}

	return result, nil
}

func UpdateUser(request *restful.Request, response *restful.Response) {
	usr := new(User)
	err := request.ReadEntity(usr)
	if err != nil {
		response.WriteErrorString(http.StatusInternalServerError, ERROR_INTERNAL)
		log.WithFields(log.Fields{"Error Msg": err}).Warn(ERROR_INTERNAL)
		return
	}
	stmt, err := db.Prepare(`
		SELECT * FROM user ORDER BY id DESC LIMIT 0, 1;`)
	if err != nil {
		response.WriteErrorString(http.StatusInternalServerError, ERROR_INTERNAL)
		log.WithFields(log.Fields{"Error Msg": err}).Warn(ERROR_INTERNAL)
		return
	}
	defer stmt.Close()
	row := stmt.QueryRow()
	var id uint64
	temp := new(User)
	err = row.Scan(&id, &temp.Name)
	if err != nil {
		response.WriteErrorString(http.StatusInternalServerError, ERROR_INTERNAL)
		log.WithFields(log.Fields{"Error Msg": err}).Warn(ERROR_INTERNAL)
		return
	}

	log.WithFields(log.Fields{"Id": id}).Debug("Highest ID")

	if usr.ID <= id {
		err := updateUser(usr)
		if err != nil {
			response.WriteErrorString(http.StatusInternalServerError, ERROR_INTERNAL)
			return
		} else {
			response.WriteEntity(true)
		}
	} else {
		CreateUser(request, response)
	}
}

func updateUser(usr *User) error {
	stmt, err := db.Prepare(`UPDATE user SET name=? WHERE id=?;`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(template.HTMLEscapeString(usr.Name), usr.ID)
	if err != nil {
		return err
	}
	return nil
}

func CreateUser(request *restful.Request, response *restful.Response) {
	usr := new(User)
	err := request.ReadEntity(usr)
	if err != nil {
		response.WriteErrorString(http.StatusBadRequest, ERROR_INVALID_INPUT)
		log.WithFields(log.Fields{"Error Msg": err}).Info(ERROR_INVALID_INPUT)
		return
	}

	id, err := createUser(usr)
	if err != nil {
		response.WriteErrorString(http.StatusInternalServerError, ERROR_INTERNAL)
		log.WithFields(log.Fields{"Error Msg": err}).Warn(ERROR_INTERNAL)
		return
	}
	response.WriteEntity("/user/" + strconv.FormatUint(id, 10))
}

func createUser(usr *User) (uint64, error) {
	stmt, err := db.Prepare(`
		INSERT INTO user ('name') VALUES (?);`)
	if err != nil {
		log.WithFields(log.Fields{"Error Msg": err}).Fatal(ERROR_STMT_PREPARE)
		return 0, err
	}
	defer stmt.Close()
	res, err := stmt.Exec(template.HTMLEscapeString(usr.Name))
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

func DeleteUser(request *restful.Request, response *restful.Response) {
	sid := request.PathParameter("id")
	id, err := strconv.ParseUint(sid, 10, 64)
	if err != nil {
		log.WithFields(log.Fields{"Error Msg": err}).Info(ERROR_INVALID_ID)
		response.WriteErrorString(http.StatusBadRequest, ERROR_INVALID_ID)
		return
	}

	err = deleteUser(id)
	if err != nil {
		log.WithFields(log.Fields{"Error Msg": err}).Info(ERROR_INTERNAL)
		response.WriteErrorString(http.StatusInternalServerError, ERROR_INTERNAL)
		return
	}
	response.WriteEntity(true)
}

func deleteUser(id uint64) error {
	stmt, err := db.Prepare("DELETE FROM user WHERE id=?;")
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
