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
	"crypto/rand"
	"crypto/sha512"
	"errors"
	"gopkg.in/mgo.v2/bson"
)

type User struct {
	ID       bson.ObjectId `bson:"_id,omitempty" json:"-"`
	Name     string        `description:"The unique identifier of a user. Let your user know that they should choose it wisely"`
	EMail    string
	Password string `bson:"-" json:",omitempty" description:"Use this field to set a new password. This field will never occour in responses."`

	Secret Secret `json:"-"`
}

type Secret struct {
	Password [sha512.Size]byte `json:"-"`
	Salt     [64]byte          `json:"-"`
}

type UserActionHistory struct {
	ItemChanges   []ItemHistory
	PolicyChanges []PolicyHistory
}

func (s *Secret) VerifyPassword(pw string) bool {
	input := s.assemblePassword(pw)
	for i := 0; i != len(s.Password); i++ {
		if s.Password[i] != input[i] {
			return false
		}
	}
	return true
}

func (s *Secret) SetPassword(pw string) error {
	err := s.genSalt()
	if err != nil {
		return err
	}

	s.Password = s.assemblePassword(pw)
	return nil
}

func (s *Secret) assemblePassword(pw string) [sha512.Size]byte {
	temp := make([]byte, len(pw)+len(s.Salt)+len(pepper))
	for i := 0; i != len(pw); i++ {
		temp[i] = pw[i]
	}
	for i := 0; i != len(s.Salt); i++ {
		temp[i+len(pw)] = s.Salt[i]
	}
	for i := 0; i != len(pepper); i++ {
		temp[i+len(pw)+len(s.Salt)] = pepper[i]
	}
	return sha512.Sum512(temp)
}

func (s *Secret) genSalt() error {
	temp := make([]byte, len(s.Salt))
	b, err := rand.Read(temp)
	if err != nil {
		return err
	}
	if b != len(s.Salt) {
		return errors.New("Read less than expected bytes")
	}
	for i := 0; i != len(s.Salt); i++ {
		s.Salt[i] = temp[i]
	}
	return nil
}

func NewUserService() *restful.WebService {
	service := new(restful.WebService)
	service.
		Path("/user").
		Doc("User related services").
		ApiVersion("0.1").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	service.Route(service.GET("/{name}").
		Param(restful.PathParameter("name", "User identifier")).
		Doc("Returns public accessable information about the referenced user").
		To(GetUserByName).
		Writes(User{}).
		Do(returnsInternalServerError, returnsNotFound, returnsBadRequest))

	service.Route(service.GET("/{name}/log").
		Param(restful.PathParameter("name", "User identifier")).
		Doc("Returns all changes the user has made in the DB").
		To(GetUserLogByName).
		Writes(UserActionHistory{}).
		Do(returnsInternalServerError, returnsNotFound, returnsBadRequest))

	service.Route(service.GET("").
		Doc("List all users (this may be replaced by a paginated version)").
		To(ListUser).
		Writes([]User{}).
		Do(returnsInternalServerError))

	service.Route(service.PUT("").
		Filter(basicAuthFilter).
		Doc("Update user information").
		To(UpdateUser).
		Reads(User{}).
		Do(returnsInternalServerError, returnsNotFound, returnsUpdateSuccessful, returnsBadRequest))

	service.Route(service.POST("").
		Doc("Register a new user").
		To(CreateUser).
		Reads(User{}).
		Returns(http.StatusOK, "Insert successful", "/user/{name}").
		Returns(http.StatusUnauthorized, "This username is not available", nil).
		Do(returnsInternalServerError, returnsBadRequest))

	service.Route(service.DELETE("/{name}").
		Filter(basicAuthFilter).
		Param(restful.PathParameter("name", "User identifier")).
		Doc("Delete a user").
		To(DeleteUser).
		Do(returnsInternalServerError, returnsNotFound, returnsDeleteSuccessful, returnsBadRequest))
	return service
}

func GetUserByName(request *restful.Request, response *restful.Response) {
	name := request.PathParameter("name")
	user, err := getUserByName(name)
	if err != nil {
		response.WriteErrorString(http.StatusBadRequest, ERROR_INVALID_ID)
		log.WithFields(log.Fields{"Error Msg": err}).
			Info(ERROR_INVALID_ID)
		return
	}
	response.WriteEntity(user)
}

func getUserByName(name string) (User, error) {
	res := User{}
	err := uCol.Find(bson.M{"name": name}).One(&res)
	return res, err
}

func GetUserLogByName(request *restful.Request, response *restful.Response) {
	name := request.PathParameter("name")
	ul, err := getUserLogByName(name)
	if err != nil {
		response.WriteErrorString(http.StatusNotFound, ERROR_INVALID_ID)
		log.Info(err)
		return
	}
	response.WriteEntity(ul)
	return
}

func getUserLogByName(name string) (*UserActionHistory, error) {
	ih := make([]ItemHistory, 0)
	err := ihCol.Find(bson.M{"user": name}).All(&ih)
	if err != nil {
		return nil, err
	}
	ph := make([]PolicyHistory, 0)
	err = phCol.Find(bson.M{"user": name}).All(&ph)
	if err != nil {
		return nil, err
	}
	ul := new(UserActionHistory)
	ul.ItemChanges = ih
	ul.PolicyChanges = ph
	return ul, nil
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
	usr := make([]User, 0)
	err := uCol.Find(nil).All(&usr)
	return usr, err
}

func UpdateUser(request *restful.Request, response *restful.Response) {
	usr := new(User)
	err := request.ReadEntity(usr)
	if err != nil {
		response.WriteErrorString(http.StatusBadRequest, ERROR_INVALID_INPUT)
		log.WithFields(log.Fields{"Error Msg": err}).Warn(ERROR_INVALID_INPUT)
		return
	}

	ex := checkUserExistance(usr)
	if ex {
		if usr.Password != "" {
			log.Debug("User supplied new password.")
			err = usr.Secret.SetPassword(usr.Password)
		} else {
			temp, err := getUserByName(usr.Name)
			if err != nil {
			} else {
				usr.Secret = temp.Secret // if no new password will be set, preserve old
			}
		}
		if err != nil { //fall through to error handling
		} else {
			err = updateUser(usr)
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

func updateUser(usr *User) error {
	return uCol.Update(bson.M{"name": usr.Name}, usr)
}

func CreateUser(request *restful.Request, response *restful.Response) {
	usr := new(User)
	err := request.ReadEntity(usr)
	log.WithFields(log.Fields{"Username": usr.Name}).Info("Attempted user registration")
	log.Debug(usr)
	if err != nil {
		response.WriteErrorString(http.StatusBadRequest, ERROR_INVALID_INPUT)
		log.WithFields(log.Fields{"Error Msg": err}).Info(ERROR_INVALID_INPUT)
		return
	}

	ex := checkUserExistance(usr)

	if ex {
		response.WriteErrorString(http.StatusUnauthorized, "This username is not available")
		return
	}
	err = usr.Secret.SetPassword(usr.Password)
	if err != nil {
	} else {

		err = createUser(usr)
	}
	if err != nil {
		response.WriteErrorString(http.StatusInternalServerError, ERROR_INTERNAL)
		log.WithFields(log.Fields{"Error Msg": err}).Warn(ERROR_INTERNAL)
		return
	}
	response.WriteEntity("/user/" + usr.Name)
}

func checkUserExistance(usr *User) bool {
	temp := User{}
	err := uCol.Find(bson.M{"name": usr.Name}).One(&temp)
	if err != nil {
		if err.Error() == "not found" {
			return false
		} else {
			log.WithFields(log.Fields{"Err": err}).
				Debug("TODO: check more error codes here")
			return true // if something strange in the neighbourhood …
			// what you gonna do? - do not register a new user!
		}
	}
	return true
}

func createUser(usr *User) error {
	return uCol.Insert(usr)
}

func DeleteUser(request *restful.Request, response *restful.Response) {
	name := request.PathParameter("name")
	log.WithFields(log.Fields{"Name": name}).Info("Got user DELETE request")
	err := deleteUser(name)
	if err != nil {
		log.WithFields(log.Fields{"Error Msg": err}).Info(ERROR_INTERNAL)
		response.WriteErrorString(http.StatusInternalServerError, ERROR_INTERNAL)
		return
	}
	response.WriteEntity(true)
}

func deleteUser(name string) error {
	return uCol.Remove(bson.M{"name": name})
}
