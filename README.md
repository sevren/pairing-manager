# Pairing Manager

This is a go microservice featuring a REST controller and a connection to RabbitMQ (Amqp)

It has 2 endpoints
* POST /pair 
* GET /code/<magic-key>

## Preqs
 * RabbitMQ running somewhere accessible (Preferably a docker container)

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

OBS!: You need to have a rabbitmq server running and fully started before you can run this code
`docker run -p 5672:5672 -p 15672:15672-it rabbitmq:3.7-management-alpine`
`go run .`

### Running with docker

`docker run --network=test3_default -it pair-man -amqp amqp://guest:guest@rabbitmq:5672/`
