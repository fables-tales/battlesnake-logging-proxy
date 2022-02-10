# battlesnake-logging-proxy

a little proxy between the internet and your battlesnake, to use, first build it:

```
go build main.go -o main
```

then migrate the database:

```
./migrate.sh
```

then run it:

`./main <bind_port> <target_host>`

where target host is the url of your battlesnake. This will dump all your game requests to games.sqlite
