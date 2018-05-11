# vaultutil [![GoDoc](https://godoc.org/github.com/DataDog/vaultutil?status.svg)](https://godoc.org/github.com/DataDog/vaultutil)

This is a Go package to make it easier to interact with vault from within Kubernetes. 

## Usage

```go
password, err = vaultutil.InClusterSecret("path/to/secret", "key")
```
