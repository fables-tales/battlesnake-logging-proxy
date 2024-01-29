package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	bind_port := os.Args[1]
	target_host := os.Args[2]
	database_file := "./games.sqlite"
	fmt.Println("Starting server on port:", bind_port)
	fmt.Println("Target host:", target_host)
	if len(os.Args) > 3 {
		database_file = os.Args[3]
	}

	db, err := sql.Open("sqlite3", database_file)
	if err != nil {
		panic("failed to open database")
	}

	defer func() {
		err = db.Close()
		if err != nil {
			panic("failed to close database")
		}
	}()

	err = db.Ping()
	if err != nil {
		panic("failed to ping database")
	}

	proxy, err := NewProxy(target_host)
	if err != nil {
		panic(err)
	}

	// handle all requests to your server using the proxy
	http.HandleFunc("/", ProxyRequestHandler(proxy, db))
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", bind_port), nil))

}

// NewProxy takes target host and creates a reverse proxy
func NewProxy(targetHost string) (*httputil.ReverseProxy, error) {
	url, err := url.Parse(targetHost)
	if err != nil {
		return nil, err
	}

	return httputil.NewSingleHostReverseProxy(url), nil
}

func insertStart(db *sql.DB, game_id string, json []byte) error {
	sqlStatement := `
	INSERT INTO starts (game_id, json)
	VALUES ($1, $2)`
	_, err := db.Exec(sqlStatement, game_id, json)
	if err != nil {
		return err
	}
	return nil
}

func insertMove(db *sql.DB, game_id string, turn int, json []byte) error {
	sqlStatement := `
	INSERT INTO moves (game_id, turn, json)
	VALUES ($1, $2, $3)`
	_, err := db.Exec(sqlStatement, game_id, turn, json)
	if err != nil {
		return err
	}
	return nil
}

func insertEnd(db *sql.DB, game_id string, json []byte) error {
	sqlStatement := `
	INSERT INTO ends (game_id, json)
	VALUES ($1, $2)`
	_, err := db.Exec(sqlStatement, game_id, json)
	if err != nil {
		return err
	}
	return nil
}

type Game struct {
	Game struct {
		Id string `json:"id"`
	} `json: "game"`
	Turn int `json:"turn"`
}

// ProxyRequestHandler handles the http request using proxy
func ProxyRequestHandler(proxy *httputil.ReverseProxy, db *sql.DB) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Proxying request:", r.URL.Path)
		origBody := r.Body
		defer origBody.Close()

		if r.URL.Path == "/start" {
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "can't read body", http.StatusBadRequest)
				return
			}
			game := Game{}
			err = json.Unmarshal(body, &game)
			if err != nil {
				http.Error(w, "can't unmarshal body", http.StatusBadRequest)
				return
			}
			r.Body = ioutil.NopCloser(bytes.NewBuffer(body))

			go insertStart(db, game.Game.Id, body)
		} else if r.URL.Path == "/move" {
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "can't read body", http.StatusBadRequest)
				return
			}
			game := Game{}
			err = json.Unmarshal(body, &game)
			if err != nil {
				http.Error(w, "can't unmarshal body", http.StatusBadRequest)
				return
			}

			r.Body = ioutil.NopCloser(bytes.NewBuffer(body))
			go func() {
				insertMove(db, game.Game.Id, game.Turn, body)
			}()
		} else if r.URL.Path == "/end" {
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "can't read body", http.StatusBadRequest)
				return
			}

			game := Game{}
			err = json.Unmarshal(body, &game)
			if err != nil {
				http.Error(w, "can't unmarshal body", http.StatusBadRequest)
				return
			}

			r.Body = ioutil.NopCloser(bytes.NewBuffer(body))
			go func() {
				insertEnd(db, game.Game.Id, body)
			}()
		}

		proxy.ServeHTTP(w, r)
	}
}
