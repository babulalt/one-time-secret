# secret-management
one time secret


# Install Dependencies
```
go mod tidy
```

# Run 
```
go run main.go
````

# Build
```
go build -o main main.go
```

To install Go Migrate

```
    $ sudo -s
    $ curl -L https://packagecloud.io/golang-migrate/migrate/gpgkey | apt-key add -
    $ echo "deb https://packagecloud.io/golang-migrate/migrate/ubuntu/ $(lsb_release -sc) main" > /etc/apt/sources.list.d/migrate.list
    $ apt-get update
    $ apt-get install -y migrate

```