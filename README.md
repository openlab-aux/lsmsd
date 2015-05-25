lsmsd - lightweight storage management system daemon [![Build Status](https://travis-ci.org/openlab-aux/lsmsd.svg?branch=master)](https://travis-ci.org/openlab-aux/lsmsd)
===========
lsmsd is a storage management system designed for our local hackerspace.
___
# Installation
lsmsd requires Go 1.4 or greater. Thanks to the Go toolset the installation is very easy. Just type the following in your terminal and press Enter:

    go get github.com/openlab-aux/lsmsd

This software needs a running instance of mongoDB. For install instructions [click here](http://docs.mongodb.org/manual/installation/)

___
# Roadmap for 0.1
  * ~~Notifications via E-Mail~~ / XMPP
  * Logging and error message overhaul
  * Better test coverage & code documentation
  * Websockets
  * Manual
  * ~~Item images (GridFS)~~

___
# Documentation

Code Documentation for this tool can be found at [godoc](http://godoc.org/github.com/openlab-aux/lsmsd).

# REST-Api Documentation

This project uses swagger to document its API.
Just run lsmsd and open http[s]://[whereitlistens]:[PORT]/apidocs/ and type http[s]://[whereitlistens]:[PORT]/apidocs.json into the textfield at the top of the page.
