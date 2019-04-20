# Pairing Manager

This is a go microservice featuring a REST controller and a connection to RabbitMQ (Amqp)

* by default it attempts to connect to the following RabbitMQ address: amqp://guest:guest@localhost:5672/ (You can override this via a command line switch `-amqp`)

This microservice is already packaged in an alpine linux docker container ready to run:

`docker run -p 8081:8081 -it sevren/pairing-manager`

The REST controller listens on port 8081

It has 2 endpoints 
* POST /pair 
* GET /{code}/{magic-key}


## Regarding Challenge 3 
This code also contains code to connect to rabbitMQ, Create an Exchange and Publish a simple JSON message. 
If for whatever reason the microservice can not connect to RabbitMQ at runtime (The container is still starting) then it will automatically
disable the RabbitMQ code.  Still leaving the REST controller available for you to use. 


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

`docker run -d -p 5672:5672 -p 15672:15672-it rabbitmq:3.7-management-alpine`

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
