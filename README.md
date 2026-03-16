# Storage Adapter Sidecar

This project is a dedicated adapter sidecar designed to bridge the gap between insecure local file uploads (specifically from tools like `ffmpeg`) and secure cloud storage clusters like MinIO or AWS S3.

## ⚠️ Architectural Requirement

**This sidecar must be run locally alongside your artifacts service.** The artifacts service cannot securely execute direct `PUT` requests to a remote, authenticated MinIO/S3 cluster on its own. This sidecar must be present on the same host/pod to securely intercept and broker these transactions.

## The Problem: FFmpeg and HLS Streaming

When using `ffmpeg` to generate HTTP Live Streaming (HLS) video, it produces a continuous stream of hundreds (or thousands) of individual `.ts` segment files. By default, `ffmpeg` attempts to upload these via basic HTTP `PUT` requests.

However, secure storage like MinIO requires authenticated requests. Generating hundreds of pre-signed `PUT` URLs in advance for every single unpredictable HLS segment is heavily inefficient and highly impractical.

## The Solution

Instead of pre-signing individual `PUT` URLs, this adapter leverages **MinIO's POST request capability with prefix keys** (allowing uploads to specific folder paths).

1. `ffmpeg` (or the artifacts service) makes a standard, unauthenticated `PUT` request to this local sidecar.
2. The sidecar intercepts the request.
3. It automatically reaches out to the main cloud drive storage service to fetch the necessary credentials and policies.
4. It translates the payload into a secure `POST` request and forwards it to the MinIO/S3 cluster.

This creates a seamless bridge, allowing tools like `ffmpeg` to function natively while keeping your storage infrastructure entirely secure.

## Configuration

By default, the sidecar application runs on `localhost:9009`, and expects the main storage server to be reachable at `127.0.0.1:8080`.

Configuration is managed natively in Go via the `internal/config/config.go` file:

```go
package config

type StorageServerConfig struct {
    AccessKey string
    Hostname  string
    Port      uint16
}

type AppConfig struct {
    Port uint16
}

type Config struct {
    MainServer StorageServerConfig
    App        AppConfig
}

func NewConfig() *Config {
    return &Config{
        MainServer: StorageServerConfig{
            AccessKey: "myAccessKey", // Main storage service API/Access key
            Hostname:  "127.0.0.1",   // Main storage service host
            Port:      uint16(8080),  // Main storage service port
        },
        App: AppConfig{
            Port: uint16(9009),      // Port this sidecar listens on
        },
    }
}

```

### Modifying Defaults

To change the ports or connection details, you can update the `NewConfig()` initializers in `internal/config/config.go` before building the sidecar, or extend this package to accept environment variables for easier containerized deployments.
