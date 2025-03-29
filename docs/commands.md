#### Install package

```bash
 go get -u $package_name
```

#### Docker compose

```bash
docker compose down && docker compose --env-file $env_file up -d --build
```

#### Generate a Swagger Doc Directory

```bash
swag init
```

#### Swagger address

```http
http://localhost:8080/swagger/index.html
```