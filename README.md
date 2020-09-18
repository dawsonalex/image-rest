# imageservice

A REST API and file server for storing and viewing images.

## Usage

CLI usage is as follows:

```
./imageservice:
  -dir string
    	the path of the directory to watch (default "/home/ad/Projects/image-rest/bin")
  -l string
    	The level of logging to use, must be one of [panic fatal error warning info debug trace] (default "info")
  -p string
    	The port to listen for API requests on. (default ":8080")
```

## Endpoints

Running `make run` will start the imageservice watching ./sample_images the following is a list of requests can responses the service accepts.

### /list

List returns the metadata about the library of images in the watch directory.

```HTTP
GET http://localhost:8080/list
```

#### Reseponse

```JSON
[
    {
        "name": "image1.jpg",
        "width": 4608,
        "height": 3456
    },
    {
        "name": "image2.jpg",
        "width": 4608,
        "height": 3456
    },
    {
        "name": "image3.jpg",
        "width": 4608,
        "height": 3456
    },
    {
        "name": "image4.jpg",
        "width": 4608,
        "height": 3456
    }
]
```

### /upload

Upload allows a user to upload images to the watch directory. The `Content-Type` of the request should be `multipart/form-data`.

The content of the request must match one of the following media types:

- image/png
- image/jpg
- image/gif



```HTTP
POST http://localhost:8080/upload
```

#### Response

If the request content doesn not match one of the stated media types, the server responds with error code 400 Bad Request.

If the request is processed without error, the server responds with 200 (OK), and an empty content body.

## Makefile

Most common usage requirements can be accomploshed using Make targets:

| Make Target | Description                                 |
|-------------|---------------------------------------------|
| build       | Build the imageservice to bin/imageservice  |
| run         | Build the imageservice and start running it |
| clean       | Clean the bin/ directory                    |
| test        | Run all tests in the service                |

