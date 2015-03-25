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
	"crypto/sha512"
	log "github.com/Sirupsen/logrus"
	"github.com/emicklei/go-restful"
	//	"gopkg.in/mgo.v2/bson"
	"net/http"
)

func basicAuthFilter(request *restful.Request, response *restful.Response, chain *restful.FilterChain) {
	_, _, ok := request.Request.BasicAuth()
	if !ok {
		response.WriteErrorString(http.StatusUnauthorized, "Parsing Error")
		log.Warn("Unsuccessful login attempt")
		return
	}
	chain.ProcessFilter(request, response)
}

func createPwHash(pw, salt string) [sha512.Size]byte {
	temp := make([]byte, len(pw)+len(salt)+len(pepper))
	for i := 0; i != len(pw); i++ {
		temp[i] = pw[i]
	}
	for i := 0; i != len(salt); i++ {
		temp[i+len(pw)] = salt[i]
	}
	for i := 0; i != len(pepper); i++ {
		temp[i+len(pw)+len(salt)] = pepper[i]
	}
	return sha512.Sum512(temp)
}
