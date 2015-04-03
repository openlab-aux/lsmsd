lsmsd - lightweight storage management system daemon
===========
lsmsd is a storage management system designed for our local hackerspace.
___
# Installation
lsmsd requires Go 1.4 or greater. Thanks to the Go toolset the installation is very easy. Just type the following in your terminal and press Enter:

    go get github.com/openlab-aux/lsmsd

___
# Roadmap for 0.1
  * Notifications via E-Mail / XMPP
  * Logging and error message overhaul
  * Better test coverage & code documentation
  * Websockets

___
# Documentation

Code Documentation for this tool can be found at [godoc](http://godoc.org/github.com/openlab-aux/lsmsd).

# REST-Api Documentation (Draft)

The following table will show you how to use this API, which errors it'll throw against you and where you can find the teapot.

Every API-Call can drop a Internal Server Error. Every API-Call which accepts JSON/XML objects in the body may drop a 400 Bad Request when the object(s) are not parsable. I won't mention it explicitly.

| Path | Method | Parameter | Description | Error Codes | Auth Required | Admin Required |
---|---|---|---|---|---|---|
| /item | GET | | List all available items (this may be replaced by a paginated version) | | | |
| | GET | id | Returns a single item identified by its ID | 404: ID Invalid | | |
| | POST | (body) Item | Insert a item into the database | | x | |
| | PUT | (body) Item | Update the item. | 404: Item not found | x | |
| | DELETE | id | Delete a item | 404: ID Invalid | x | |
| /user | GET | | Return details of the requesting user | | x | |
| | POST | (body) User | Register a new user | 401: This username is not available | | |
| | PUT | (body) User | Update user information | 404: User not found. Please register first | x | (x) |
| | DELETE | | Delete the User | | x | (x) |
| /policies | GET | | List all available policies (this may be replaced by a paginated version | | | |
| | GET | id | Returns a single policy identified by its ID | 404: ID Invalid | | | |
| | POST | (body) Policy | Insert a new policy | | x | |
| | PUT | (body) Policy | Update / Insert a policy | | x | |
| | DELETE | id | Delete a policy | | x | |
