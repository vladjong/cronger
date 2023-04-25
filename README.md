![Cronger](logo_cronger.png)

# Cronger

`cronger` is an expressive work planning package with permanent data storage based on the package [go-co-op/gocron](https://github.com/go-co-op/gocron)

## Installation

If using go modules.

```sh
go get -u github.com/vladjong/cronger
```

## Quick Examples

### Config

Basic configuration `cronger`

```go
type Config struct {
	Loc        *time.Location // time zone (time.UTC)
	TypeClient int            // type client (Sqlx)
	Client     interface{}    // driver client (*sqlx.DB)
	IsMigrate  bool           // whether the jobs table has been created (true)
}
```

### New

Creating an instance `cronger`

*When you restart, a backup of old jobs is created*

```go
cr, err := cronger.New(&cronger.Config{
	Loc:        time.UTC,
	TypeClient: cronger.Sqlx,
	Client:     db,
	IsMigrate:  true,
})
```

### Stop

Stop performing jobs `cronger`  

```go
jobs, err := cr.Jobs()
```

### StartAsync

Starts `cronger` asynchronously

```go
cr.StartAsync()
```

### Add

Add a new job to `cronger`

*If a job is in a backup, it is removed from the backup*

```go
err := cr.AddJob("title", "tag", "* * * * *", func() {
	fmt.Println("done task!")
})
```

### Remove

Remove a job to `cronger` in tag

*If a job is in a backup, it is removed from the backup*

```go
err := cr.RemoveJob("tag2")

```

### GetJobs

List of active jobs

```go
jobs, err := cr.Jobs()
```

### GetBackupJobs

List of jobs since the last restart service

```go
backJob = cr.BackupJobs()
```

For more examples, take a look in our [examples](example/sqlx_example/main.go)

## Supported drivers

- [x] Sqlx
- [ ] Mongo
- [ ] MySql

## Running tests

```sh
go test -v -race ./...
```
