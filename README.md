# Pasgan

Pasgan (Scots Gaelic for "package") is a tool that analyzes saved Docker images and generates a Dockerfile that could have been used to create that image.

## Installation

```
go install github.com/raesene/pasgan@latest
```

## Usage

Save a Docker image:

```
docker save nginx:latest -o nginx.tar
```

Analyze the image and generate a Dockerfile:

```
pasgan analyze nginx.tar
```

Output to a file:

```
pasgan analyze nginx.tar -o Dockerfile
```

## Features

- Extracts and analyzes Docker image metadata
- Reconstructs Dockerfile instructions from image layers
- Handles multi-stage builds (coming soon)
- Identifies base images
- Reconstructs RUN, COPY, ENV, EXPOSE, etc. commands

## Requirements

- Go 1.18 or higher

## License

MIT