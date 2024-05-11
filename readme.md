
## FileKeeper
FileKeeper is a file hosting service built with pure Golang and some JavaScript. It allows users to securely upload, store, and manage files on a server.


## Environment Variables

To run this project, you will need to add the following environment variables to your .env file [example file](https://gist.github.com/fossyy/a15e6806cd33d0ee74611ff410c9d1aa)

`SERVER_HOST`

`SERVER_PORT`

`DOMAIN`

`CORS_PROTO`

`CORS_LIST`

`CORS_METHODS`

`DB_HOST`

`DB_PORT`

`DB_USERNAME`

`DB_PASSWORD`

`DB_NAME`

`SMTP_HOST`

`SMTP_PORT`

`SMTP_USER`

`SMTP_PASSWORD`

`SESSION_NAME`

`SESSION_MAX_AGE`

## Usage
#### Running with Docker

1. Ensure you have Docker installed on your system. If not, download and install Docker from [here](https://docs.docker.com/get-docker/). 

2. Clone the repository:
```bash
  docker pull fossyy/filekeeper:latest
```

3. Install dependencies:
```bash
  docker run -d --name filekeeper -v /path/to/data:/src/uploads -p 8000:8000 fossyy/filekeeper:latest
```

4. Access the application in your web browser:
http://localhost:8080

#### Running Locally
If you prefer to run the app locally without Docker, follow these steps:
1. Clone the repository:
```bash
  https://github.com/fossyy/filekeeper.git
```

2. Install dependencies::
```bash
  cd filekeeper
  go mod tidy
```

3. Build and run the application:
```bash
  make
```

4. Access the application in your web browser:
http://localhost:8080




## Demo

https://filekeeper.my.id


## Authors

- [@Bagas Aulia Rezki (fossyy)](https://github.com/fossyy)

## License
[Creative Commons Attribution–NonCommercial 4.0 CC BY-NC](https://creativecommons.org/licenses/by-nc-nd/4.0)
