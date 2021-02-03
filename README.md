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

## List all images in the library

List images that are in the directory that the image service is watching.

```HTTP
GET /list
```

### Response

```HTTP
Status: 200 OK
```

```JSON
[
    {
        "name": "image1.jpg",
        "width": 4608,
        "height": 3456
    }
]
```

## Upload an image to the library

Upload an image or multiple images to the watch directory. The `Content-Type` of the request should be `multipart/form-data`.

The content of the request must match one of the following media types:

- image/png
- image/jpg
- image/gif

```HTTP
POST /upload
```

#### Response

If the request content does not match one of the stated media types, the server responds with a status of `415 Unsupported Media Type` and empty body.

If the request is processed without error, the server responds with `200 OK`, and an empty body.

## Remove a file from the library

Remove an image from the watch directory by name.

```HTTP
DELETE /remove
```

### Parameters 

| **Name** | **Type** | **Description** |
|----------|----------|-----------------|
| `name`   | `string` | **Required** The name of the image. |

### Response

The `/remove` endpoint always responds with an empty body, but the status is different under different circumstances:

If the file is removed:

```HTTP
Status: 200 OK
```

If the filename doesn't exist:

```HTTP
Status: 404 Not Found
```

## Get an image from the library

Get an image from the library by name:

```HTTP
GET /image
```

### Parameters

| **Name** | **Type** | **Description** |
|----------|----------|-----------------|
| `name`   | `string` | **Required** The name of the image. |

If the `name` parameter is passed multiple times, the first instance is used.

### Response

If an image exists:

```HTTP
Status: 200 OK
Content-Type: image/*
```

and content containing an images bytes.

If the image doesn't exist:

```HTTP
Status: 404 Not Found
```

If the image name is not valid (The `name` param should only contain the image name, no directory prefix is necessary):

```HTTP
Status: 400 Bad Request
```


## Makefile

Most common usage requirements can be accomplished using Make targets:

| Make Target | Description                                 |
|-------------|---------------------------------------------|
| build       | Build the imageservice to bin/imageservice  |
| run         | Build the imageservice and start running it |
| clean       | Clean the bin/ directory                    |
| test        | Run all tests in the service                |

