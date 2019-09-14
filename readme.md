# Jessica Mock Tool

Jessica is a simulator for HTTP-based APIs. Some might consider it a service virtualization tool or a mock server.

It enables you to stay productive when an API you depend on doesn't exist or ins't complete. 
It supports testing of edge cases and failure modes that the real API won't reliable produce.
And because it's fast it can reduce your build time from hours down to minutes. 

# Basic structure

```
├── static
│   ├── auth.json
│   └── login.html
├── jessica.json (configuration file)
└── jessica (binary file)
```

# Jessica Config file

```json
{
  "version": "0.2",
  "port": "5000",
  "allowed_headers": "Content-Type, X-CSRF-Token, Authorization, access-control-expose-headers",
  "allowed_origins": "*",
  "allowed_methods": "GET, HEAD, POST, PUT, OPTIONS",
  "stubs": [
    {
      "request": {
        "url": "/api/authenticate",
        "method": "POST"
      },
      "response": {
        "status": 401,
        "content": "auth.json"
      }
    },
    {
      "request": {
        "url": "/login.html",
        "method": "GET"
      },
      "response": {
        "content": "login.html",
        "content-type": "text/html"
      }
    }
  ]
}
```

# Building Jessica Mock locally

To build jessica

`go build -o bin/jessica`

To run jessica

`go run main.go`
