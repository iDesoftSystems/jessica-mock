# Jessica Mock Tool

Jessica is a simulator for HTTP-based APIs. Some might consider it a service virtualization tool or a mock server.

It enables you to stay productive when an API you depend on doesn't exist or ins't complete. 
It supports testing of edge cases and failure modes that the real API won't reliable produce.
And because it's fast it can reduce your build time from hours down to minutes. 

# Basic structure

|--static
|----login.json
|----user.json
|--jessica.json (configuration file)
|--jessica (binary file)

# Jessica Config file

```json
{
  "version": "0.1",
  "port": "5000",
  "routes": [
    {
      "method": "GET",
      "path": "/login",
      "data": "login.json"
    },
    {
      "method": "GET",
      "path": "/user",
      "data": "user.json"
    }
  ]
}
```

# Building Jessica Mock locally

To build jessica

`go build -o jessica`

To run jessica

`go run main.go`
