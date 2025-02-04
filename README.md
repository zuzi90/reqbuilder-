# reqbuilder

`reqbuilder` is a helper package for sending HTTP requests in tests. It simplifies the process of making requests, handling headers, cookies, and processing responses.

## Features

- Supports `POST`, `GET`, and other HTTP methods
- Handles `multipart/form-data` requests
- Manages cookies automatically
- Supports gzip, Brotli, zstd, and deflate response decompression
- Uses `testify/require` for assertions in tests

## Installation

```sh
go get github.com/zuzi90/reqbuilder-
```

## Usage

### Creating a Request Builder

```go
import (
    "context"
    "testing"
    "github.com/stretchr/testify/require"
    "your-repo/reqbuilder"
)

func TestExample(t *testing.T) {
    builder := reqbuilder.New(require.New(t))
    ctx := context.Background()

    response, cookies := builder.Request(
        t, ctx, "POST", "https://example.com", "/api/login",
        []byte(`{"username": "test", "password": "pass"}`),
        nil, nil, "Bearer token")
    
    require.Equal(t, 200, response.StatusCode)
    _ = cookies
}
```

### Sending Multipart Requests

```go
response, cookies := builder.MultipartRequest(
    t, ctx, "POST", "https://example.com", "/upload",
    []byte("file content"), "json", nil, nil, "Bearer token")
```

### Sending Requests Without a Body

```go
response, cookies := builder.RequestWithoutBody(
    t, ctx, "GET", "https://example.com", "/profile", nil, nil, "Bearer token")
```

### Reading Response Body

```go
body, err := builder.ReadResponseBody(response)
require.NoError(t, err)
```





