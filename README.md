goCeiling
---------

This is a system for manage drawings of ceilings. 

At this moment supports only REST API and SQLite storage. 
For more read [README](/api/README.md) in API directory.

If your computer has installed Golang, you can install application with the next commands:
```shell script
go get github.com/maxsid/goCeilings
go install github.com/maxsid/goCeilings
```

Usage of goCeilings command:
```
  -addr string
        address of API for listening (default "127.0.0.1:8081")
  -admin
        create a new administrator (the app create an admin user automatically if database doesn't have at least one)
  -salt string
        salt for users passwords (default value in the code)
  -secret string
        secret for singing of jwt token (default value in the code)
  -sqlite string
        file of SQLite data storage (default "go-ceiling.db")
```