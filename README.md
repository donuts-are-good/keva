![keva logo](https://github.com/donuts-are-good/keva/assets/96031819/fea4e301-122d-473a-8478-5ee8fb3f735f)
![donuts-are-good's followers](https://img.shields.io/github/followers/donuts-are-good?&color=555&style=for-the-badge&label=followers) ![donuts-are-good's stars](https://img.shields.io/github/stars/donuts-are-good?affiliations=OWNER%2CCOLLABORATOR&color=555&style=for-the-badge) ![donuts-are-good's visitors](https://komarev.com/ghpvc/?username=donuts-are-good&color=555555&style=for-the-badge&label=visitors)
# keva

keva is a key:value datastore with http json interface.
## usage

build it:

```
go mod tidy

go build
```

run it the normal way:
```
./keva
```
if you're fancy:
```
./keva --port=:8080 --savepath=mydata.json --saveinterval=10s 
```


### flags:

- `--port` (optional): define the server port.
    
    - default: `:8080`
- `--savepath` (optional): define the path to save the key-value store data.
    
    - default: `data.json`
- `--saveinterval` (optional): define the interval to automatically save data.
    
    - default: `5s`


## demo:

to better understand the usage of `keva`, here's a series of `curl` commands to try it out on your own:

### 1. store a key-value:



`curl -x post http://localhost:8080/store/demokey -h "content-type: application/json" -d '{"value": "demo value"}'`

**output:**

`ok`

### 2. retrieve a stored value:



`curl -x get http://localhost:8080/store/demokey`

**output:**



`"demo value"`

### 3. delete a stored key:



`curl -x delete http://localhost:8080/store/demokey`

**output:**

`ok`

### 4. try retrieving a deleted key:



`curl -x get http://localhost:8080/store/demokey`

**output:**


`key not found`

### 5. health check:



`curl -x get http://localhost:8080/health`

**output:**

`healthy`

## endpoints:

### 1. get /store/{key}

**description:** retrieve a value by the given key.

**parameters:**

- `key`: the key associated with the stored value.

**response:**

- `200 ok`: value retrieved successfully. it returns the stored value in json format.
- `404 not found`: key not found.

**example:**



`curl -x get http://localhost:8080/store/examplekey`

---

### 2. post /store/{key}

**description:** store a value associated with the given key.

**parameters:**

- `key`: the key to store the value with.

**request body:** json object containing the value to store.

- `value` (string): the value to be stored.

**response:**

- `201 created`: key-value set successfully.
- `400 bad request`: no value provided or bad request format.

**example:**



`curl -x post http://localhost:8080/store/examplekey -h "content-type: application/json" -d '{"value": "this is an example value"}'`

---

### 3. delete /store/{key}

**description:** delete the value associated with the given key.

**parameters:**

- `key`: the key of the value to delete.

**response:**

- `200 ok`: key deleted successfully.
- `404 not found`: key not found.

**example:**



`curl -x delete http://localhost:8080/store/examplekey`

---

### 4. get /health

**description:** health check endpoint.

**response:**

- `200 ok`: healthy.

**example:**



`curl -x get http://localhost:8080/health`

## errors:

the api uses conventional http response codes to indicate the success or failure of an api request.

- `200 ok`: the request was successful.
- `201 created`: the request was successful and a resource was created.
- `400 bad request`: the request could not be understood or was missing required parameters.
- `404 not found`: resource not found. this can be used when a specific key does not exist in the store.
- `405 method not allowed`: the http method used is not valid for the specific endpoint.

## concurrency and data persistence:

- the keyvalue store is thread-safe, using mutexes to ensure that concurrent requests do not conflict.
- the data is periodically saved to a file (default path is `data.json`) at a set interval (default is every 5 seconds). this behavior ensures that the data remains persistent even if the application restarts.

## license

mit license 2023 donuts-are-good, for more info see license.md
