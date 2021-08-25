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
	mrand "math/rand"
	"net/http"
)

const (
	ERROR_INVALID_ID    = "Error: Invalid ID"
	ERROR_STMT_PREPARE  = "Error: Statement prepare failed"
	ERROR_INVALID_INPUT = "Error: Invalid Input"
	ERROR_INTERNAL      = "Error: Internal Server Error"
	ERROR_INSERT        = "Error: DB Insert failed"
	ERROR_QUERY         = "Error: DB Query failed"
)

func DebugLoggingFilter(rq *restful.Request, rs *restful.Response, ch *restful.FilterChain) {
	id := uint32(mrand.Int31())
	log.WithFields(log.Fields{
		"ID": id, "Path": rq.SelectedRoutePath()}).
		Debug("Got Request")

	log.WithFields(log.Fields{
		"ID": id, "PathParameters": rq.PathParameters()}).
		Debug()

	log.WithFields(log.Fields{
		"ID": id, "Method": rq.Request.Method}).
		Debug()

	log.WithFields(log.Fields{
		"ID": id, "Protocol": rq.Request.Proto}).
		Debug()

	log.WithFields(log.Fields{
		"ID": id, "Host": rq.Request.Header.Get("Host")}).
		Debug()

	log.WithFields(log.Fields{
		"ID": id, "Upgrade": rq.Request.Header.Get("Upgrade")}).
		Debug()

	log.WithFields(log.Fields{
		"ID": id, "User-Agent": rq.Request.Header.Get("User-Agent")}).
		Debug()

	log.WithFields(log.Fields{
		"ID": id, "Content-Length": rq.Request.ContentLength}).
		Debug()

	ch.ProcessFilter(rq, rs)
}

func returnsInternalServerError(b *restful.RouteBuilder) {
	b.Returns(http.StatusInternalServerError, ERROR_INTERNAL, nil)
}

func returnsNotFound(b *restful.RouteBuilder) {
	b.Returns(http.StatusNotFound, ERROR_INVALID_ID, nil)
}

func returnsUpdateSuccessful(b *restful.RouteBuilder) {
	b.Returns(http.StatusOK, "Update successful", nil)
}

func returnsDeleteSuccessful(b *restful.RouteBuilder) {
	b.Returns(http.StatusOK, "Delete successful", nil)
}

func returnsBadRequest(b *restful.RouteBuilder) {
	b.Returns(http.StatusBadRequest, "Failed to parse input", nil)
}
