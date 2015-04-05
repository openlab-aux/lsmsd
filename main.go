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
	Mail    backend.Mailconfig
	Logging struct {
		Level string
	}
}

const (
	defaultConfigPath        = "./config.gcfg"
	defaultNetworkAddress    = ":8080"
	defaultLogLevel          = "INFO"
	defaultCrypto            = false
	defaultCertificate       = "./cert.pem"
	defaultKeyfile           = "keyfile.key"
	defaultDatabaseServer    = "localhost"
	defaultDatabase          = "lsmsd"
	defaultPepperfile        = "./.pepper"
	defaultMailEnabled       = false
	defaultMailStartTLS      = true
	defaultMailServerAddress = ""
	defaultMailUsername      = ""
	defaultMailPassword      = ""
	defaultMailEMailAddress  = ""
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

	var mailenabled = flag.Bool("enablemail", defaultMailEnabled, "enable email notifications")
	var mailstarttls = flag.Bool("mailtls", defaultMailStartTLS, "use TLS when sending emails")
	var mailserver = flag.String("mailserver", defaultMailServerAddress, "address and port of your smtp server")
	var mailuser = flag.String("mailuser", defaultMailUsername, "smtp username")
	var mailpassword = flag.String("mailpassword", defaultMailPassword, "smtp password")
	var mailaddress = flag.String("mailaddress", defaultMailEMailAddress, "email address")

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

	if *mailenabled != defaultMailEnabled {
		cfg.Mail.Enabled = *mailenabled
	}
	if *mailstarttls != defaultMailStartTLS {
		cfg.Mail.StartTLS = *mailstarttls
	}
	if *mailserver != defaultMailServerAddress {
		cfg.Mail.ServerAddress = *mailserver
	}
	if *mailuser != defaultMailUsername {
		cfg.Mail.Username = *mailuser
	}
	if *mailpassword != defaultMailPassword {
		cfg.Mail.Password = *mailpassword
	}
	if *mailaddress != defaultMailEMailAddress {
		cfg.Mail.EMailAddress = *mailaddress
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

	log.Debug(cfg)

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
