# sqlite-rest

<p align="center">
  <a href="https://github.com/jonamat/sqlite-rest/actions">
    <img alt="GitHub Workflow Status (branch)" src="https://img.shields.io/github/actions/workflow/status/jonamat/sqlite-rest/docker-image.yml" />
  </a>

  <a href="https://github.com/jonamat/sqlite-rest/blob/master/go.mod">
    <img alt="GitHub go.mod Go version" src="https://img.shields.io/github/go-mod/go-version/jonamat/sqlite-rest" />
  </a>

  <a href="https://hub.docker.com/r/jonamat/sqlite-rest">
    <img alt="Docker Image Size (tag)" src="https://img.shields.io/docker/image-size/jonamat/sqlite-rest/latest" />
  </a>
</p>

Expose CRUD operations for SQLite database over HTTP via REST API. 

## Installation

### From releases page
  
Download the binary for your platform from the [releases page](https://github.com/jonamat/sqlite-rest/releases)

### From source

```bash
$ go get github.com/jonamat/sqlite-rest
```

### From Docker

```bash
$ docker pull jonamat/sqlite-rest
```

## CLI usage

```bash
$ sqlite-rest -h
Usage of sqlite-rest:
  -f string
        Path to the SQLite database file (default "db.sqlite")
  -p string
        Port to listen on (default 8080)

# Example with default values
$ sqlite-rest 
```

## Docker usage

Only 6 MB of size. Built available for ARM64, ARMv7 and AMD64.

```bash
$ docker run -p 8080:8080 -v /path/to/db.sqlite:/data.sqlite jonamat/sqlite-rest

# Or, if you want to use a different name for the database file
$ docker run -p 8080:8080 -v /path/to/db.sqlite:/db.sqlite jonamat/sqlite-rest -f /db.sqlite
```

## API
[Search all](#search-all)<br>
[Get record by id](#get-record-by-id)<br>
[Create record](#create-record)<br>
[Update record](#update-record)<br>
[Delete record](#delete-record)<br>
[Exec](#exec)<br>

### Search all

Get all record in a table.<br>

Request: `GET /:table`<br>

Basic example:<br>

```bash
$ curl localhost:8080/cats

{
  "data": [
    { "id": 1, "name": "Tequila", "paw": 4 },
    { "id": 2, "name": "Whisky", "paw": 3 }
  ],
  "limit": null,
  "offset": null,
  "total_rows": 2
}

```

Optional parameters:<br>

- `offset`: Offset the number of records returned. Default: `0`
- `limit`: Limit the number of records returned. Default: not set
- `order_by`: Order the records by a column. Default: `id`
- `order_dir`: Order the records by a column. Default: `asc`
- `filters_raw`: Filter the records by a raw SQL query. Must be URIescaped.
- `filters`: Filter the records by a JSON object. Must be URIescaped.

Filters:<br>

Can be passed as a JSON object or as a raw WHERE clause. The JSON object is more convenient to use, the raw query is more flexible. Both must be URIescaped. Cannot be used together.

Example with `filters_raw` parameter in cURL:<br>

```bash
$ curl "localhost:8080/cats?offset=10&limit=2&order_by=name&order_dir=desc&filters_raw=paw%20%3D%204%20OR%20name%20LIKE%20'%25Tequila%25'"

{
  "data": [
    { "id": 10, "name": "Tequila", "paw": 4 },
    { "id": 11, "name": "Cognac", "paw": 4 }
  ],
  "limit": 2,
  "offset": 10,
  "total_rows": 2
}
```

Example with `filters_raw` parameter in JS:<br>

```js
const filters = "paw = 4 OR name LIKE '%Tequila%'"

fetch(`http://localhost:8080/cats?filters_raw=${encodeURIComponent(filters)}`)
```

Example with `filters` parameter in JS:<br>

```js
const filters = [
  {
    "column": "paw",
    "operator": "=",
    "value": 4
  },
  {
    "column": "name",
    "operator": "LIKE",
    "value": "%Tequila%"
  }
]

fetch(`http://localhost:8080/cats?filters=${encodeURIComponent(JSON.stringify(filters))}`)
```

### Get record by id

Get a record by its id in a table.<br>

Request: `GET /:table/:id`<br>

Example:<br>

```bash
$ curl localhost:8080/cats/1

{
  "data": { 
    "id": 1, 
    "name": 
    "Tequila", 
    "paw": 4 
  }
}
```

### Create record

Create a record in a table.<br>

Request: `POST /:table`<br>

Example:<br>

```bash
$ curl -X POST -H "Content-Type: application/json" -d '{"name": "Tequila", "paw": 4}' localhost:8080/cats

{
  "id": 1,
}
```

### Update record

Update a record in a table.<br>
⚠️ The update is a PATCH, not a PUT. Only the fields passed in the body will be updated. The other fields will be left untouched.

Request: `PATCH /:table/:id`<br>

Example:<br>

```bash
$ curl -X PATCH -H "Content-Type: application/json" -d '{"name": "Tequila", "paw": 4}' localhost:8080/cats/1

{
  "id": 1,
}
```

### Delete record

Delete a record in a table.<br>

Request: `DELETE /:table/:id`<br>

Example:<br>

```bash
$ curl -X DELETE localhost:8080/cats/1

{
  "id": 1,
}
```

### Exec

Execute an arbitrary query. ⚠️ Experimental<br>

Request: `OPTIONS /__/exec`<br>

Example:<br>

```bash
$ curl -X OPTIONS -H "Content-Type: application/json" -d '{"query": "create table cats (id PRIMARY_KEY, name TEXT, paw INTEGER)"}' localhost:8080/__/exec

{
  "status": "success", 
}
```

## License

MIT