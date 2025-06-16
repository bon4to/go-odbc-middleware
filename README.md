# CDE API - IBM DB2 Query Service

A lightweight HTTP service built in Go to execute SQL queries against IBM DB2 databases using the [go\_ibm\_db](https://github.com/ibmdb/go_ibm_db) driver.

## 📦 Requirements

- Go 1.20+
- IBM DB2 CLI Driver installed (for macOS: `dsdriver`)
- `.env` file with database credentials

## 🧪 .env Structure

Create a `.env` file in the root directory with the following variables:

```env
DB_HOST=your-db-host
DB_PORT=50000
DB_USER=your-db-user
DB_PASSWORD=your-password
DB_DSN_1=your-database-name-1
DB_DSN_2=your-database-name-2
```

You can add more `DB_DSN_n` entries if needed. The client must specify the `source` index.

## 🔧 Build Instructions

### 🖥️ macOS (native build)

```bash
go build -o cde_api main.go
```

### 🪟 Cross-compilation for Windows

You need to set `GOOS` and `GOARCH`:

```bash
GOOS=windows GOARCH=amd64 go build -o cde_api.exe main.go
```

> ⚠️ The IBM DB2 driver must be available in the build machine. For Windows, you may need to compile inside a Windows VM or Docker container with the proper environment.

## 🚀 Running the Server

```bash
./cde_api
```

The server will start on port `40500` and expose one endpoint:

## 📡 Endpoint

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

## ⚠️ Security Warning

This API executes raw SQL. This is dangerous if exposed to untrusted users.

### Best Practices:

- Restrict usage via **HMAC or token-based authentication**.
- Validate and sanitize inputs.
- Use read-only DB users.
- Consider limiting SQL to whitelisted queries only.

## 📄 License

MIT License.

## 🧠 Author

Maintained by bon4to.

