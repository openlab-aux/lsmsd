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

package main

import (
	"flag"
	//	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/emicklei/go-restful"
	"github.com/openlab-aux/lsmsd/backend"
	"net/http"
	//	"time"
	"code.google.com/p/gcfg"
	"github.com/emicklei/go-restful/swagger"
	"gopkg.in/mgo.v2"
)

type Config struct {
	Network struct {
		ListenTo string
	}
	Crypto struct {
		Enabled     bool
		Certificate string
		KeyFile     string
		Pepperfile  string
	}
	Database struct {
		Server string
		DB     string
	}
	Logging struct {
		Level string
	}
}

const (
	defaultConfigPath     = "./config.gcfg"
	defaultNetworkAddress = ":8080"
	defaultLogLevel       = "INFO"
	defaultCrypto         = false
	defaultCertificate    = "./cert.pem"
	defaultKeyfile        = "keyfile.key"
	defaultDatabaseServer = "localhost"
	defaultDatabase       = "lsmsd"
	defaultPepperfile     = "./.pepper"
)

func main() {
	log.WithFields(log.Fields{"Version": "0.1"}).Info("lsmsd starting")
	var cfg Config
	var configpath = flag.String("cfgpath", defaultConfigPath, "path to your config file")
	var listento = flag.String("listento", defaultNetworkAddress, "listen to address and port")
	var loglevel = flag.String("loglevel", defaultLogLevel, "verbosity")
	var enableCrypto = flag.Bool("enablecrypto", defaultCrypto, "Use TLS instead of plain text")
	var certificate = flag.String("certificate", defaultCertificate, "certificate path")
	var pepper = flag.String("pepper", defaultPepperfile, "path to your pepperfile")
	var keyfile = flag.String("keyfile", defaultKeyfile, "private key path")
	var dbserver = flag.String("dbserver", defaultDatabaseServer, "address of your mongo db server")
	var dbdb = flag.String("dbdb", defaultDatabase, "default database name")
	flag.Parse()
	err := gcfg.ReadFileInto(&cfg, *configpath)

	if err != nil {
		log.Fatal(err)
	}

	if *listento != defaultNetworkAddress {
		cfg.Network.ListenTo = *listento
	}
	if *loglevel != defaultLogLevel {
		cfg.Logging.Level = *loglevel
	}
	if *enableCrypto != defaultCrypto {
		cfg.Crypto.Enabled = *enableCrypto
	}
	if *certificate != defaultCertificate {
		cfg.Crypto.Certificate = *certificate
	}
	if *keyfile != defaultKeyfile {
		cfg.Crypto.KeyFile = *keyfile
	}
	if *dbserver != defaultDatabaseServer {
		cfg.Database.Server = *dbserver
	}
	if *dbdb != defaultDatabase {
		cfg.Database.DB = *dbdb
	}
	if *pepper != defaultPepperfile {
		cfg.Crypto.Pepperfile = *pepper
	}

	switch cfg.Logging.Level {
	case "Debug":
		log.SetLevel(log.DebugLevel)
	}

	// Test DB Connection
	log.Info("Test database connection …")
	s, err := mgo.Dial(cfg.Database.Server)
	if err != nil {
		log.Fatal(err)
	}
	log.Info("Connection successful")
	defer s.Close()

	backend.RegisterDatabase(s, cfg.Database.DB)
	backend.ReadPepper(cfg.Crypto.Pepperfile)
	defer backend.CloseIDGen()
	restful.DefaultContainer.Filter(restful.DefaultContainer.OPTIONSFilter)
	restful.Add(backend.NewItemService())
	restful.Add(backend.NewUserService())
	restful.Add(backend.NewPolicyService())

	if log.GetLevel() == log.DebugLevel {
		restful.DefaultContainer.Filter(backend.DebugLoggingFilter)
	}

	config := swagger.Config{
		WebServices: restful.DefaultContainer.RegisteredWebServices(),
		//WebServicesUrl:  "", //cfg.Network.ListenTo,
		ApiPath:         "/apidocs.json",
		SwaggerPath:     "/apidocs/",
		SwaggerFilePath: "./swagger/dist/",
	}

	swagger.RegisterSwaggerService(config, restful.DefaultContainer)

	log.WithFields(log.Fields{"Address": cfg.Network.ListenTo, "TLS": cfg.Crypto.Enabled}).
		Info("lsms started successfully")
	if cfg.Crypto.Enabled {
		log.Fatal(http.ListenAndServeTLS(cfg.Network.ListenTo, cfg.Crypto.Certificate, cfg.Crypto.KeyFile, nil))
	} else {
		log.Fatal(http.ListenAndServe(cfg.Network.ListenTo, nil))
	}
}
