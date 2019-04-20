package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/render"
	"github.com/sevren/pair-man/cache"
	"github.com/sevren/pair-man/rabbit"
	log "github.com/sirupsen/logrus"
)

// The request from the /pair post will be mapped to this
// Warning there is validation on the the `device` so it must be a proper ipv4 or ipv6 address
type PostReq struct {
	Code   string `json:"code"`
	Device net.IP `json:"device"`
}

// The response for the /pair endpoint
type PairPayload struct {
	Key string `json:"key"`
}

// The rabbitMQ message
type Msg struct {
	Code string `json:"code"`
}

type ErrorPayload struct {
	Error ErrorResponse `json:"error"`
}

type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type SuccessResponse struct {
	Message string
}

// Rest Controller - Sets up routes required
func Routes(conn *rabbit.RMQConn) (*chi.Mux, error) {

	// Sets up a new in memory cache - If this was a real microservice we would probably use Redis for a proper cache..
	inMemCache := cache.New()

	// Worst part about working with web stuff is CORS :X .. This should setup a reasonable CORS configuration
	corsConf := corsConfig()

	r := chi.NewRouter()
	r.Use(corsConf.Handler)
	r.Use(middleware.RealIP) // We use a middleware RealIP to be able to process the header X-FORWARDED-FOR
	r.Use(render.SetContentType(render.ContentTypeJSON))
	r.Use(middleware.Logger)

	// Provides the Rest routes
	// /pair and /pair/{code}/{magic-key}
	r.Route("/pair", func(r chi.Router) {

		r.Post("/", func(w http.ResponseWriter, r *http.Request) {
			render.SetContentType(render.ContentTypeJSON)
			decoder := json.NewDecoder(r.Body)
			var p PostReq
			err := decoder.Decode(&p)
			if err != nil {
				errPayload := ErrorPayload{ErrorResponse{Code: http.StatusBadRequest, Message: err.Error()}}
				render.Status(r, http.StatusBadRequest)
				render.JSON(w, r, errPayload)
				return
			}

			// Get the servers current time, We work with Unix (seconds) from Epoch
			currTime := time.Now()
			exp := currTime.Add(20 * time.Second).Unix() // Set the expiry time

			// Generates the magic key from server time and formats it to just HHmm
			magickey := currTime.Format("1504") // Welcome to go's horrible way of extracting datetime stuff ... :( https://golang.org/src/time/format.go

			item := cache.Codes{Code: p.Code, IP: p.Device.String(), Created: currTime.Unix(), Expiration: exp}

			// Cache lookup key is a combo of Code:Magickey
			cacheKey := fmt.Sprintf("%s:%s", p.Code, magickey)

			inMemCache.Insert(cacheKey, item)

			// Challenge 3 - Assuming RabbitMQ server is up and the connection is valid..
			if conn != nil {
				rmqMsg := Msg{}
				rmqMsg.Code = item.Code
				log.Infof("\nPublishing message to %s, %+v", conn.Ex, rmqMsg)
				payload, err := json.Marshal(rmqMsg)
				if err != nil {
					log.Fatalf("%s: %s", "Failed to marshal JSON", err)
				}
				conn.PublishMessage(payload)
			}

			render.JSON(w, r, PairPayload{magickey})

		})

		r.Route("/{code}", func(r chi.Router) {
			r.Get("/{magic-key}", func(w http.ResponseWriter, r *http.Request) {
				code := chi.URLParam(r, "code")
				magickey := chi.URLParam(r, "magic-key")

				cacheKey := fmt.Sprintf("%s:%s", code, magickey)

				if !inMemCache.Exists(cacheKey) {
					errPayload := ErrorPayload{ErrorResponse{Code: http.StatusNotFound, Message: "Pairing does not exist"}}
					render.Status(r, http.StatusNotFound)
					render.JSON(w, r, errPayload)
					return
				}

				cachedItem := inMemCache.Get(cacheKey)
				currTime := time.Now().Unix()

				// Check the IP address of the client, if the request comes from a different address then disallow
				if r.RemoteAddr != cachedItem.IP {
					errPayload := ErrorPayload{ErrorResponse{Code: http.StatusForbidden, Message: "Code rejected, - Requesting ip address not correct"}}
					render.Status(r, http.StatusForbidden)
					render.JSON(w, r, errPayload)
					return
				}

				// Check the expiry time on the magic-key
				// if > 1 hour then return not found and clean the cache
				if inMemCache.IsExpired(cacheKey, currTime) {
					errPayload := ErrorPayload{ErrorResponse{Code: http.StatusNotFound, Message: "magic-key has expired, please pair again"}}
					render.Status(r, http.StatusNotFound)
					render.JSON(w, r, errPayload)
					inMemCache.Delete(cacheKey)
					return
				}
				render.JSON(w, r, SuccessResponse{Message: "success"})

			})
		})

	})

	return r, nil
}

func corsConfig() *cors.Cors {
	return cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"POST", "GET", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Content-Type", "X-CSRF-Token", "Cache-Control", "X-Requested-With", "X-Forwarded-For"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	})
}
