# Commonly used commands

## Install package

```bash
 go get -u $package_name
```

## Docker compose

```bash
docker compose down && docker compose -f $docker_compose_file --env-file $env_file up -d --build
```

## Docker build image

```bash
docker build -t "$IMAGE_NAME" .
```

## Generate a Swagger Doc Directory

```bash
swag init
```

## Swagger address

```http
http://localhost:8080/swagger/index.html
```

## Run tests

```bash
go test $test_dir -v
```
