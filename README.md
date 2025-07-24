# go-odbc-middleware


![badge-db2](https://img.shields.io/badge/IBM%20DB2-supported-blue)
![badge-sqlite](https://img.shields.io/badge/SQLite3-supported-blue)
![badge-sqlite](https://img.shields.io/badge/PostgreSQL-upcoming-yellow)
![badge-sqlite](https://img.shields.io/badge/MySQL-upcoming-yellow)
![badge-sqlite](https://img.shields.io/badge/SQL%20Server-upcoming-yellow)

**A lightweight, concurrent microservice API built in Go for executing SQL queries against multiple ODBC drivers.**

---

## Requirements

- Golang 1.20+

- `.env` file with database credentials

  > Duplicate `.env.example` to `.env` and fill in the values.

- IBM DB2 CLI Driver installed

## How to build

### Cross-compilation for Windows

You need to set `GOOS` and `GOARCH`:

```bash
GOOS=windows GOARCH=amd64 go build -o cde_api.exe main.go
```

> ⚠️ The IBM DB2 driver must be available in the build machine. For Windows, you may need to compile inside a Windows VM or Docker container with the proper environment.

## Running the Server

```bash
./cde_api.exe
```

The server will start on port `40500` and expose one endpoint.

## Running as a Service (Recommended)

This application can (and should) be installed as a system service for production use.

### Windows (with NSSM)

1. [Download NSSM](https://nssm.cc/release/nssm-2.24.zip) and add it to PATH
2. Run:

    ```powershell
    nssm install CDE_API_Service "C:\cde-api\cde_api.exe"
    ```

3. Start the service:

    ```powershell
    nssm start CDE_API_Service
    ```


> Make sure the port is free (e.g. 40500) and your environment variables are set correctly.

---

## Endpoint

### `POST /query`

#### Request Body (JSON):

```json
{
  "query": "SELECT * FROM table_name",
  "source": 1
}
```

- `query`: SQL string to execute
- `source`: integer matching the `DB_DSN_n` environment variable

#### Response:

```json
{
  "columns": ["id", "name", "email"],
  "data": [
    {"id": 1, "name": "Alice", "email": "alice@example.com"},
    {"id": 2, "name": "Bob", "email": "bob@example.com"}
  ]
}
```

> ⚠️ This API executes raw SQL. Do not expose to untrusted users.

### Upcoming

- Token-based auth (e.g. HMAC or JWT)
- Configurable query whitelisting
- Logging control and access auditing
- Rate limiting to prevent abuse

Feel free to contribute or open issues for feature requests.

## 

This app is maintained by [@lucas-bonato](https://www.github.com/lucas-bonato).
