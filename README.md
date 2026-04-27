# tyk-sre-assignment

This repository contains the boilerplate projects for the SRE role interview assignments.

### Project

Location: https://github.com/TykTechnologies/tyk-sre-assignment/tree/main/golang

In order to build the project run:
```
go mod tidy & go build
```

To run it against a real Kubernetes API server:
```
./tyk-sre-assignment --kubeconfig '/path/to/your/kube/conf' --address ":8080"
```

To execute unit tests:
```
go test -v
```

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


