# Docker-compose for forensic purposes

Set up required containers via:

```bash
docker-compose up --build
```

This will expose:

* `lsmsd-frickelclient` on Port 5000
* `lsmsd` on Port 8080

In order to create the default user required for Frickelclient:

```bash
http -v POST http://localhost:8080/users "Name=Inventur" "Password=Inventur"
```
