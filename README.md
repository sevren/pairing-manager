# Pairing Manager

This is a go microservice featuring a REST controller and a connection to RabbitMQ (Amqp)

* by default it attempts to connect to the following RabbitMQ address: amqp://guest:guest@localhost:5672/ (You can override this via a command line switch `-amqp`)

This microservice is already packaged in an alpine linux docker container ready to run:

`docker run -p 8081:8081 -it sevren/pairing-manager`

## Regarding Challenge 3 

This code also contains code to connect to rabbitMQ, Create an Exchange and Publish a simple JSON message. 
If for whatever reason the microservice can not connect to RabbitMQ at runtime (The container is still starting) then it will automatically
disable the RabbitMQ code. Still leaving the REST controller available for you to use. 

If you want to checkout challenge 3 stuff, please start the rabbitmq server with the following 

`docker run -d -p 5672:5672 -p 15672:15672 -it rabbitmq:3.7-management-alpine`

You can log into the RabbitMQ server by opening a web-browser and navigating to the following:

http://localhost:15672
* Login with guest/guest

Click on Exchanges
Click on data

If the message has been sent properly you should see a spike in the graph for messages in a few seconds (may have to refresh the page)
![alt text](./imgs/post-req.png "PostReq")

## REST endpoints

The REST controller listens on port 8081

It has 2 endpoints 
* POST /pair 
* GET /{code}/{magic-key}

## Regarding the POST endpoint 
The REST endpoint /pair allows you to get a magic-key produced from the license in license-manager microservice. 
The code will ingest a post body of `{"code":"x123y", "device":"<IP-ADDR>"}` 
* code is the license from challenge 1
* device is the ip address from the client

*OBS: There is validation on the device input as it MUST be a ip address type either IPV4 or IPV6*

You can run a quick test using the following:

`curl -d '{"code":"x123y", "device":"192.168.1.1"}' http://localhost:8081/pair`

Output

```json
{"key":"1346"}
```
### RabbitMQ Communication

If the microservice is connected to RabbitMQ, then making a POST will trigger a publish to the RabbitMQ exchange called `data`
The message that will be sent looks like this:
```json
{"code":"x123y"}
```


## Regarding the GET endpoint

The input is the Code you used in the last request and the magic-key that was produced from the POST /pair endpoint

In order to get a successful pairing, you need to call the GET /pair/{code}/{magic-key} with the same ip address that was set in your POST. 
If you are testing to see if the GET will be rejected from a diffrent ip address you can spoof it by using the following header `X-Forwarded-For` for the request.

`curl --header "X-Forwarded-For: 192.168.1.1" http://localhost:8081/x123y/1346`

Output

```json
{
    "Message": "success"
}
```

Assuming the magic-key has not yet expired you can use following to test if the code is rejected if you try to GET from the wrong ip-address

`curl --header "X-Forwarded-For: 192.168.1.255" http://localhost:8081/x123y/1346`

Output

```json
{
    "error": {
        "code": 403,
        "message": "Code rejected, - Requesting ip address not correct"
    }
}
```

## Preqs 
 *  Challenge 3 - The RabbitMQ server needs to be running somewhere accessible (Preferably a docker container)

## Building
To build from source you require the following: 
* Go (1.12)
* Make
* Docker (With the ability to run without sudo..)

You can build this locally by using Make

`make static`

### Docker Container
The following will create a local docker image with the dockertag set

`make DOCKERTAG=pair-man docker`

You can view your images by 
`docker image`

## Running

To run locally you can use the following. 

If you wish to test the service to service communication, you need to run a rabbit mq server.
The following will start one in a docker container.

*OBS!: RabbitMQ is notorious for taking it's time to start up, please give the container a minute or so to be fully booted*

`docker run -d -p 5672:5672 -p 15672:15672 -it rabbitmq:3.7-management-alpine`

To run (with default AMQP) the microservice from code simply:

`go run .`

### Overriding the RabbitMQ url

You can override the rabbitmq url by giving the command line flag `-amqp <RMQ URL>`
The format of the <RMQ URL> is "amqp://{user}:{password}@<host>:<port>/"

### Running with docker

If you have build this image locally: 

`docker run --network=host -it pair-man -amqp amqp://guest:guest@localhost:5672/`

If you wish to just use my already published docker image from dockerhub: 

`docker run --network=host -it sevren/pairing-manager`
