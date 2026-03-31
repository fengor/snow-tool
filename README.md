# snow-tool

A Go CLI tool for interacting with the [ServiceNow Table API](https://www.servicenow.com/docs/r/api-reference/rest-apis/c_TableAPI.html). Compiles to a single binary with **no external dependencies**.

## Build

```bash
# Windows binary (cross-compile from any platform)
GOOS=windows GOARCH=amd64 go build -o snow-tool.exe .

# Native build
go build -o snow-tool .
```

## Configuration

Set the following environment variables before running:

| Variable        | Description                               |
|-----------------|-------------------------------------------|
| `SNOW_INSTANCE` | Instance URL, e.g. `dev12345.service-now.com`       |
| `SNOW_USER`     | Basic-auth username                       |
| `SNOW_PASSWORD` | Basic-auth password                       |

```bat
set SNOW_INSTANCE=dev12345
set SNOW_USER=admin
set SNOW_PASSWORD=yourpassword
```

## Usage

```
snow-tool <table> <verb> [args...]
```

### Verbs

| Verb     | Arguments  | Description                        |
|----------|------------|------------------------------------|
| `list`   | _(none)_   | List all records in the table      |
| `get`    | `<sys_id>` | Get a single record by sys_id      |
| `create` | `<json>`   | Create a record from a JSON string |
| `delete` | `<sys_id>` | Delete a record by sys_id          |

### Table Keywords

Short aliases for common tables — or pass any raw ServiceNow table name directly:

| Keyword(s)          | Table            |
|---------------------|------------------|
| `incident`, `inc`   | `incident`       |
| `change`, `chg`     | `change_request` |
| `problem`, `prb`    | `problem`        |
| `user`, `usr`       | `sys_user`       |
| `group`, `grp`      | `sys_user_group` |
| `ci`, `cmdb`        | `cmdb_ci`        |
| `task`              | `task`           |
| `request`, `req`    | `sc_request`     |
| `ritm`              | `sc_req_item`    |
| `catalog`           | `sc_cat_item`    |
| `knowledge`, `kb`   | `kb_knowledge`   |

Any unrecognised keyword is passed through verbatim (e.g. `cmdb_ci_server`).

## Examples

```bat
REM List all incidents
snow-tool.exe incident list

REM Get a single incident by sys_id
snow-tool.exe inc get 1234abcd1234abcd1234abcd1234abcd

REM Create an incident
snow-tool.exe inc create "{\"short_description\":\"Disk full\",\"urgency\":\"1\"}"

REM Delete an incident
snow-tool.exe inc delete 1234abcd1234abcd1234abcd1234abcd

REM Use a raw table name (no keyword mapping needed)
snow-tool.exe cmdb_ci_server list
```

## Dependencies

None — uses only the Go standard library.
