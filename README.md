# DPacks Storage Microservice

## Description
This microservice is responsible for storing and retrieving files from the DPack storage. It is a simple REST API that allows users to upload files to the storage and retrieve them later.

## Installation
1. Clone the repository
2. Run `docker build -t dpacks-storage .` to build the Docker image
3. Run `docker run -p 4004:4004 dpacks-storage` to start the container
4. The service will be available at `http://localhost:4004`

## Endpoints
- `POST /upload`: Upload a file to the storage
- - **Request Body:** `multipart/form-data` with the file to be uploaded as `file` and the filename as `filename`
- - **Response:** `200 OK` with the file ID

## Technologies
- Go
- Gin
- Docker
- Google Firebase Storage

## Copyright
&copy; 2024 DPacks. All Rights Reserved.