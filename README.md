# Warden

`warden` is a Kubernetes cluster health monitoring tool built on top of the [tyk-sre-assignment](https://github.com/TykTechnologies/tyk-sre-assignment/tree/main/golang) boilerplate as part of the Tyk SRE assignment.

It monitors deployment replica drift across all namespaces and ensures the tool is always connected to the Kubernetes API server.

## Endpoints

| Endpoint | Description |
|---|---|
| `/healthz` | Health check for Warden itself |
| `/deploymentHealth` | Lists deployments where available replicas don't match desired |
| `/apiServerConnectivity` | Verifies connectivity to the Kubernetes API server |

## Prerequisites
- Docker
- A valid kubeconfig at `~/.kube/config`

### Running with Docker

To pull the image from GHCR
```
docker pull ghcr.io/shruddhabhat/tyk-sre-assignment/warden:main
```

To run the image
```
docker run --network host \
  -v ~/.kube/config:/app/kubeconfig \
  ghcr.io/shruddhabhat/tyk-sre-assignment/warden:main \
  /app/warden --kubeconfig /app/kubeconfig
```

### Development

Build
```
go mod tidy && go build
```

Test
```
go test -v
```

Run locally
```
./warden --kubeconfig '/path/to/your/kube/conf' --address ":8080"
```



