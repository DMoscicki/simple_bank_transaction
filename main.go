package main

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"sync"
	"time"
)

type userRequest struct {
	Values map[string]int32
	Cha    chan struct{}
	Locker sync.RWMutex
}

type Client struct {
	UUID  int32  `json:"uuid"`
	Count int32  `json:"count"`
	Name  string `json:"name"`
}

func (req *userRequest) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	switch r.Method {
	case http.MethodPost:
		req.Post(w, r)
	default:
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid method"))
	}
}

func (req *userRequest) Post(w http.ResponseWriter, r *http.Request) {

	log.Println(r.Method)

	from := r.Header.Get("From")
	val := r.Header.Get("Count")
	target := r.Header.Get("To")

	log.Println(target)

	if from == "" || val == "" || target == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("empty one of the header param"))
	}

	log.Println(val)

	diff, err := strconv.Atoi(val)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("error"))
	}

	log.Println(int32(diff))

	req.Locker.RLock()
	defer req.Locker.RUnlock()
	value, ok := req.Values[from]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
	}
	targetval, ok := req.Values[target]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
	}

	if diff > 0 && int32(diff) <= value {
		req.Values[from] = value - int32(diff)
		req.Values[target] = targetval + int32(diff)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("успех"))
	} else {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Введите другое значение"))
	}
}

func main() {

	var wg sync.WaitGroup

	mux := http.NewServeMux()

	newlst, err := fillMap()
	if err != nil {
		panic(err)
	}

	log.Println(newlst)

	mux.Handle("/transfer", &userRequest{
		Values: newlst,
		Cha:    make(chan struct{}, 0),
	})

	go func() {
		ticker := time.NewTicker(time.Minute * 1)
		for _ = range ticker.C {
			wg.Add(2)
			log.Println("Parallel update")
			go makebackupdb(&wg)
			go insertData(newlst, &wg)
			wg.Wait()
		}
	}()

	go func() {
		log.Fatal(http.ListenAndServe(":8080", mux))
	}()
	log.Println("server started")

	stopC := make(chan os.Signal)
	signal.Notify(stopC, os.Interrupt)
	<-stopC

	wg.Add(2)

	go makebackupdb(&wg)
	go insertData(newlst, &wg)

	wg.Wait()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	log.Println("server stopping...")
	defer cancel()

	<-ctx.Done()

}

func fillMap() (map[string]int32, error) {
	newlst := make(map[string]int32, 0)

	conn, err := newDB()
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = conn.Close(context.TODO())
	}()

	rows, err := conn.Query(context.Background(), "select * from balance natural join client")
	if err != nil {
		log.Println("error query")
		return nil, err
	}

	defer func() {
		rows.Close()
	}()

	for rows.Next() {
		var data Client
		err = rows.Scan(&data.UUID, &data.Count, &data.Name)
		if err != nil {
			return nil, err
		}
		newlst[data.Name] = data.Count
	}

	return newlst, nil
}

func insertData(newlst map[string]int32, wg *sync.WaitGroup) {
	defer wg.Done()
	conn, err := newDB()
	if err != nil {
		log.Fatal(err)
	}
	tx, err := conn.Begin(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		_ = tx.Rollback(context.Background())
	}()

	for k, v := range newlst {
		_, err = tx.Exec(context.Background(), "update balance set count = $1 where uuid = (SELECT uuid FROM client WHERE name = $2);", v, k)
		if err != nil {
			log.Fatal(err)
		}
	}

	_ = tx.Commit(context.Background())
}

func newDB() (*pgx.Conn, error) {
	conn, err := pgx.Connect(context.Background(), "postgres://dmitrij:password@localhost:5432/clients")
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func makebackupdb(wg *sync.WaitGroup) {
	defer wg.Done()
	cmd := exec.Command("docker", "exec", "-i", "client", "/bin/bash", "-c", `PGPASSWORD=password pg_dump --username dmitrij clients`, ">", "db/dump.sql")

	cmd.Stderr = os.Stderr
	stdout, err := cmd.Output()

	if err != nil {
		log.Println(err)
	}

	f, err := os.Create("db/dump.sql")
	if err != nil {
		log.Fatal(err)
	}
	_, err = f.Write(stdout)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("backup done")
}
