# vaultutil

This is a go package to make it easier to interact with vault from with Kubernetes.

## Usage

```go
password, err = vaultutil.InClusterSecret("path/to/secret", "key")
```
