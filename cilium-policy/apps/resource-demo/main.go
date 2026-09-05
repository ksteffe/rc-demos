package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/nats-io/nats.go"
	"github.com/redis/go-redis/v9"
)

type application struct {
	db             *sql.DB
	redis          *redis.Client
	nats           *nats.Conn
	inventoryURL   string
	fulfillmentURL string
}

type resourceRequest struct {
	ID          string    `json:"id"`
	ItemID      string    `json:"itemId"`
	Quantity    int       `json:"quantity"`
	Destination string    `json:"destination"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"createdAt"`
}

type reservationInput struct {
	RequestID string `json:"requestId"`
	ItemID    string `json:"itemId"`
	Quantity  int    `json:"quantity"`
}

type fulfillmentInput struct {
	RequestID   string `json:"requestId"`
	ItemID      string `json:"itemId"`
	Quantity    int    `json:"quantity"`
	Destination string `json:"destination"`
}

func main() {
	role := env("SERVICE_ROLE", "request")
	app := &application{}
	var handler http.Handler
	var err error

	switch role {
	case "request":
		handler, err = app.requestHandler()
	case "inventory":
		handler, err = app.inventoryHandler()
	case "fulfillment":
		handler, err = app.fulfillmentHandler()
	case "notification":
		err = app.runNotificationWorker()
		if err == nil {
			return
		}
	case "directory":
		handler = app.directoryHandler()
	default:
		err = fmt.Errorf("unknown SERVICE_ROLE %q", role)
	}
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.Handle("/", handler)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	server := &http.Server{
		Addr:              ":8080",
		Handler:           requestLog(mux),
		ReadHeaderTimeout: 5 * time.Second,
	}
	go func() {
		log.Printf("%s service listening on :8080", role)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = server.Shutdown(ctx)
}

func (a *application) requestHandler() (http.Handler, error) {
	db, err := openDatabase()
	if err != nil {
		return nil, err
	}
	a.db = db
	a.inventoryURL = env("INVENTORY_URL", "http://stock-provider-v2.applications.svc.cluster.local:8080")
	a.fulfillmentURL = env("FULFILLMENT_URL", "http://dispatch-planner-v1.applications.svc.cluster.local:8080")
	redisAddress := env("REDIS_ADDR",
		env("REDIS_HOST", "availability-cache-v1.dependencies.svc.cluster.local")+":"+env("REDIS_PORT", "6379"))
	a.redis = redis.NewClient(&redis.Options{Addr: redisAddress})
	a.nats, _ = nats.Connect(env("NATS_URL", "nats://event-bus-nats.dependencies.svc.cluster.local:4222"),
		nats.Timeout(2*time.Second), nats.MaxReconnects(1))

	if err := retryDatabase(db, `
		CREATE TABLE IF NOT EXISTS requests (
			id text PRIMARY KEY,
			item_id text NOT NULL,
			quantity integer NOT NULL,
			destination text NOT NULL,
			status text NOT NULL,
			created_at timestamptz NOT NULL
		)`); err != nil {
		return nil, err
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /requests", a.createRequest)
	mux.HandleFunc("GET /requests/{id}", a.getRequest)
	mux.HandleFunc("GET /", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = io.WriteString(w, requestForm)
	})
	return mux, nil
}

func (a *application) createRequest(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ItemID      string `json:"itemId"`
		Quantity    int    `json:"quantity"`
		Destination string `json:"destination"`
	}
	if err := decodeJSON(r, &input); err != nil || input.ItemID == "" || input.Quantity < 1 || input.Destination == "" {
		writeError(w, http.StatusBadRequest, "itemId, a positive quantity, and destination are required")
		return
	}

	id := fmt.Sprintf("req-%d", time.Now().UnixNano())
	ctx, cancel := context.WithTimeout(r.Context(), 4*time.Second)
	defer cancel()

	availabilityURL := a.inventoryURL + "/items/" + input.ItemID + "/availability"
	var availability struct {
		Available int `json:"available"`
	}
	if err := getJSON(ctx, availabilityURL, &availability); err != nil {
		writeError(w, http.StatusServiceUnavailable, "inventory capability is unreachable: "+err.Error())
		return
	}
	if availability.Available < input.Quantity {
		writeError(w, http.StatusConflict, "insufficient stock")
		return
	}

	reservation := reservationInput{RequestID: id, ItemID: input.ItemID, Quantity: input.Quantity}
	if err := postJSON(ctx, a.inventoryURL+"/reservations", reservation, nil); err != nil {
		writeError(w, http.StatusServiceUnavailable, "reservation capability failed: "+err.Error())
		return
	}

	req := resourceRequest{
		ID: id, ItemID: input.ItemID, Quantity: input.Quantity,
		Destination: input.Destination, Status: "scheduled", CreatedAt: time.Now().UTC(),
	}
	if _, err := a.db.ExecContext(ctx,
		`INSERT INTO requests (id,item_id,quantity,destination,status,created_at) VALUES ($1,$2,$3,$4,$5,$6)`,
		req.ID, req.ItemID, req.Quantity, req.Destination, req.Status, req.CreatedAt); err != nil {
		writeError(w, http.StatusServiceUnavailable, "could not persist request: "+err.Error())
		return
	}
	if err := postJSON(ctx, a.fulfillmentURL+"/fulfillment-tasks", fulfillmentInput{
		RequestID: id, ItemID: input.ItemID, Quantity: input.Quantity, Destination: input.Destination,
	}, nil); err != nil {
		writeError(w, http.StatusServiceUnavailable, "fulfillment capability failed: "+err.Error())
		return
	}

	encoded, _ := json.Marshal(req)
	if err := a.redis.Set(ctx, "request:"+id, encoded, 10*time.Minute).Err(); err != nil {
		writeError(w, http.StatusServiceUnavailable, "could not cache request: "+err.Error())
		return
	}
	if a.nats != nil && a.nats.IsConnected() {
		_ = a.nats.Publish("fulfillment.requested", encoded)
	}
	writeJSON(w, http.StatusCreated, req)
}

func (a *application) getRequest(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()
	if cached, err := a.redis.Get(ctx, "request:"+id).Bytes(); err == nil {
		w.Header().Set("X-Data-Source", "redis")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(cached)
		return
	}
	var result resourceRequest
	err := a.db.QueryRowContext(ctx,
		`SELECT id,item_id,quantity,destination,status,created_at FROM requests WHERE id=$1`, id).
		Scan(&result.ID, &result.ItemID, &result.Quantity, &result.Destination, &result.Status, &result.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		writeError(w, http.StatusNotFound, "request not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (a *application) inventoryHandler() (http.Handler, error) {
	db, err := openDatabase()
	if err != nil {
		return nil, err
	}
	a.db = db
	if err := retryDatabase(db, `
		CREATE TABLE IF NOT EXISTS inventory (
			item_id text PRIMARY KEY,
			available integer NOT NULL CHECK (available >= 0)
		)`); err != nil {
		return nil, err
	}
	if err := retryDatabase(db, `
		CREATE TABLE IF NOT EXISTS reservations (
			request_id text PRIMARY KEY,
			item_id text NOT NULL,
			quantity integer NOT NULL,
			created_at timestamptz NOT NULL DEFAULT now()
		)`); err != nil {
		return nil, err
	}
	if err := retryDatabase(db, `
		INSERT INTO inventory(item_id,available) VALUES
			('blankets',100),('water-filters',50),('first-aid-kits',25)
		ON CONFLICT (item_id) DO NOTHING`); err != nil {
		return nil, err
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /items/{itemId}/availability", a.getAvailability)
	mux.HandleFunc("POST /reservations", a.createReservation)
	mux.HandleFunc("POST /admin/items", func(w http.ResponseWriter, _ *http.Request) {
		writeError(w, http.StatusNotImplemented, "admin operation intentionally outside the profile")
	})
	return mux, nil
}

func (a *application) getAvailability(w http.ResponseWriter, r *http.Request) {
	var available int
	err := a.db.QueryRowContext(r.Context(), `SELECT available FROM inventory WHERE item_id=$1`, r.PathValue("itemId")).Scan(&available)
	if errors.Is(err, sql.ErrNoRows) {
		writeError(w, http.StatusNotFound, "item not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"itemId": r.PathValue("itemId"), "available": available})
}

func (a *application) createReservation(w http.ResponseWriter, r *http.Request) {
	var input reservationInput
	if err := decodeJSON(r, &input); err != nil || input.RequestID == "" || input.ItemID == "" || input.Quantity < 1 {
		writeError(w, http.StatusBadRequest, "requestId, itemId, and a positive quantity are required")
		return
	}
	tx, err := a.db.BeginTx(r.Context(), nil)
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, err.Error())
		return
	}
	defer tx.Rollback()
	result, err := tx.ExecContext(r.Context(),
		`UPDATE inventory SET available=available-$1 WHERE item_id=$2 AND available >= $1`,
		input.Quantity, input.ItemID)
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, err.Error())
		return
	}
	changed, _ := result.RowsAffected()
	if changed == 0 {
		writeError(w, http.StatusConflict, "insufficient stock")
		return
	}
	if _, err = tx.ExecContext(r.Context(),
		`INSERT INTO reservations(request_id,item_id,quantity) VALUES($1,$2,$3)`,
		input.RequestID, input.ItemID, input.Quantity); err != nil {
		writeError(w, http.StatusConflict, "request is already reserved")
		return
	}
	if err = tx.Commit(); err != nil {
		writeError(w, http.StatusServiceUnavailable, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"requestId": input.RequestID, "status": "reserved"})
}

func (a *application) fulfillmentHandler() (http.Handler, error) {
	db, err := openDatabase()
	if err != nil {
		return nil, err
	}
	a.db = db
	if err := retryDatabase(db, `
		CREATE TABLE IF NOT EXISTS fulfillment_tasks (
			id text PRIMARY KEY,
			request_id text UNIQUE NOT NULL,
			item_id text NOT NULL,
			quantity integer NOT NULL,
			destination text NOT NULL,
			status text NOT NULL,
			created_at timestamptz NOT NULL DEFAULT now()
		)`); err != nil {
		return nil, err
	}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /fulfillment-tasks", a.createFulfillmentTask)
	mux.HandleFunc("GET /fulfillment-tasks/{id}", a.getFulfillmentTask)
	return mux, nil
}

func (a *application) createFulfillmentTask(w http.ResponseWriter, r *http.Request) {
	var input fulfillmentInput
	if err := decodeJSON(r, &input); err != nil || input.RequestID == "" || input.Destination == "" {
		writeError(w, http.StatusBadRequest, "requestId and destination are required")
		return
	}
	id := fmt.Sprintf("task-%d", time.Now().UnixNano())
	_, err := a.db.ExecContext(r.Context(), `
		INSERT INTO fulfillment_tasks(id,request_id,item_id,quantity,destination,status)
		VALUES($1,$2,$3,$4,$5,'queued')`,
		id, input.RequestID, input.ItemID, input.Quantity, input.Destination)
	if err != nil {
		writeError(w, http.StatusConflict, "task already exists or could not be stored")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"id": id, "requestId": input.RequestID, "status": "queued"})
}

func (a *application) getFulfillmentTask(w http.ResponseWriter, r *http.Request) {
	var id, requestID, status string
	err := a.db.QueryRowContext(r.Context(),
		`SELECT id,request_id,status FROM fulfillment_tasks WHERE id=$1`, r.PathValue("id")).
		Scan(&id, &requestID, &status)
	if errors.Is(err, sql.ErrNoRows) {
		writeError(w, http.StatusNotFound, "task not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"id": id, "requestId": requestID, "status": status})
}

func (a *application) runNotificationWorker() error {
	db, err := openDatabase()
	if err != nil {
		return err
	}
	if err := retryDatabase(db, `
		CREATE TABLE IF NOT EXISTS notifications (
			request_id text PRIMARY KEY,
			payload jsonb NOT NULL,
			recorded_at timestamptz NOT NULL DEFAULT now()
		)`); err != nil {
		return err
	}
	nc, err := nats.Connect(env("NATS_URL", "nats://event-bus-nats.dependencies.svc.cluster.local:4222"),
		nats.RetryOnFailedConnect(true), nats.MaxReconnects(-1))
	if err != nil {
		return err
	}
	defer nc.Close()
	_, err = nc.Subscribe("fulfillment.requested", func(msg *nats.Msg) {
		var request resourceRequest
		if json.Unmarshal(msg.Data, &request) != nil {
			return
		}
		if _, err := db.Exec(`
			INSERT INTO notifications(request_id,payload) VALUES($1,$2)
			ON CONFLICT (request_id) DO NOTHING`, request.ID, msg.Data); err != nil {
			log.Printf("record notification: %v", err)
		}
	})
	if err != nil {
		return err
	}
	log.Print("notification worker subscribed to fulfillment.requested")
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	return nil
}

func (a *application) directoryHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /donors/{donorId}", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"id": r.PathValue("donorId"), "status": "active"})
	})
	mux.HandleFunc("POST /donors/search", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, []any{})
	})
	return mux
}

func openDatabase() (*sql.DB, error) {
	db, err := sql.Open("pgx", env("DATABASE_URL",
		"postgres://resources:local-demo-only@resource-records-v1.dependencies.svc.cluster.local:5432/resources?sslmode=disable"))
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(5)
	return db, nil
}

func retryDatabase(db *sql.DB, statement string) error {
	var err error
	for attempt := 1; attempt <= 30; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		_, err = db.ExecContext(ctx, statement)
		cancel()
		if err == nil {
			return nil
		}
		log.Printf("database unavailable (attempt %d/30): %v", attempt, err)
		time.Sleep(2 * time.Second)
	}
	return fmt.Errorf("database did not become ready: %w", err)
}

func getJSON(ctx context.Context, url string, output any) error {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return fmt.Errorf("%s: %s", resp.Status, strings.TrimSpace(string(body)))
	}
	return json.NewDecoder(resp.Body).Decode(output)
}

func postJSON(ctx context.Context, url string, input, output any) error {
	body, _ := json.Marshal(input)
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return fmt.Errorf("%s: %s", resp.Status, strings.TrimSpace(string(body)))
	}
	if output != nil {
		return json.NewDecoder(resp.Body).Decode(output)
	}
	return nil
}

func decodeJSON(r *http.Request, target any) error {
	defer r.Body.Close()
	return json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(target)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func env(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}

func requestLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start).Round(time.Millisecond))
	})
}

const requestForm = `<!doctype html>
<html lang="en"><head><meta charset="utf-8"><title>Community resource requests</title></head>
<body><main><h1>Request community resources</h1>
<p>Submit JSON to <code>POST /requests</code> with itemId, quantity, and destination.</p>
<pre>curl -X POST http://localhost:8080/requests -H 'content-type: application/json' \
-d '{"itemId":"blankets","quantity":2,"destination":"Shelter A"}'</pre>
</main></body></html>`
