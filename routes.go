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

type PostReq struct {
	Code   string `json:"code"`
	Device net.IP `json:"device"`
}

type PairPayload struct {
	Key string `json:"key"`
}

type Msg struct {
	Code string `json:"code"`
}

func Routes(conn *rabbit.RMQConn) (*chi.Mux, error) {

	c := cache.New()
	corsConf := corsConfig()
	r := chi.NewRouter()
	r.Use(corsConf.Handler)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)

	r.Route("/pair", func(r chi.Router) {

		r.Post("/", func(w http.ResponseWriter, r *http.Request) {

			decoder := json.NewDecoder(r.Body)
			var p PostReq
			err := decoder.Decode(&p)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			realIP := r.RemoteAddr
			fmt.Println(realIP)

			t := time.Now()
			exp := t.Add(20 * time.Second).Unix() // Set the expiry time

			// Generates the magic key from server time and formats it to just HHmm
			magickey := t.Format("1504") // Welcome to go's horrible way of extracting datetime stuff ... :( https://golang.org/src/time/format.go
			fmt.Println("p.Device.String(): ", p.Device.String())
			item := cache.Codes{Code: p.Code, IP: p.Device.String(), Created: t.Unix(), Expiration: exp}
			c.Insert(magickey, item)

			// Challenge 3 - Assuming RabbitMQ server is up and the connection is valid..
			if conn != nil {
				rmqMsg := Msg{}
				rmqMsg.Code = item.Code
				log.Infof("Publishing message to %s, %+v", conn.Ex, rmqMsg)
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

				magickey := chi.URLParam(r, "magic-key")
				i := c.Get(magickey)
				t := time.Now().Unix()

				// Check the IP address of the client, if the request comes from a different address then disallow
				// if the key does not exist in the cache you also get forbidden since i.IP will be empty
				if r.RemoteAddr != i.IP {
					w.WriteHeader(http.StatusForbidden)
					render.JSON(w, r, PairPayload{"Code rejected, - Requesting address not correct"})
					return
				}

				// Check the expiry time on the magic-key
				// if > 1 hour then return not found and clean the cache
				if c.IsExpired(magickey, t) {
					w.WriteHeader(http.StatusNotFound)
					render.JSON(w, r, PairPayload{"expired"})
					c.Delete(magickey)
					return
				}
				render.JSON(w, r, PairPayload{"success"})

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
