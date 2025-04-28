package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sync"
	"testing"
	"time"

	"git.adyxax.org/adyxax/tfstated/pkg/database"
)

var baseURI = url.URL{
	Host:   "127.0.0.1:8082",
	Path:   "/",
	Scheme: "http",
}
var db *database.DB
var adminPassword string
var adminPasswordMutex sync.Mutex

func TestMain(m *testing.M) {
	getenv := func(key string) string {
		switch key {
		case "TFSTATED_DATA_ENCRYPTION_KEY":
			return "hP3ZSCnY3LMgfTQjwTaGrhKwdA0yXMXIfv67OJnntqM="
		case "TFSTATED_HOST":
			return "127.0.0.1"
		case "TFSTATED_PORT":
			return "8082"
		case "TFSTATED_SESSIONS_SALT":
			return "a528D1m9q3IZxLinSmHmeKxrx3Pmm7GQ3nBzIDxjr0A="
		case "TFSTATED_VERSIONS_HISTORY_LIMIT":
			return "3"
		default:
			return ""
		}
	}

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	_ = os.Remove("./test.db")
	var err error
	db, err = database.NewDB(ctx, "./test.db", getenv)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(1)
	}
	database.AdvertiseAdminPassword = func(password string) {
		adminPasswordMutex.Lock()
		defer adminPasswordMutex.Unlock()
		adminPassword = password
	}
	go run(
		ctx,
		db,
		getenv,
	)
	err = waitForReady(ctx, 5*time.Second, "http://127.0.0.1:8082/healthz")
	if err != nil {
		fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(1)
	}

	ret := m.Run()

	cancel()
	err = db.Close()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(1)
	}
	_ = os.Remove("./test.db")

	os.Exit(ret)
}

func runHTTPRequest(method string, auth bool, uriRef *url.URL, body io.Reader, testFunc func(*http.Response, error)) {
	uri := baseURI.ResolveReference(uriRef)
	client := http.Client{}
	req, err := http.NewRequest(method, uri.String(), body)
	if err != nil {
		testFunc(nil, fmt.Errorf("failed to create request: %w", err))
		return
	}
	if auth {
		adminPasswordMutex.Lock()
		defer adminPasswordMutex.Unlock()
		req.SetBasicAuth("admin", adminPassword)
	}
	resp, err := client.Do(req)
	if err != nil {
		testFunc(nil, fmt.Errorf("failed to do request: %w\n", err))
		return
	}
	testFunc(resp, nil)
	_ = resp.Body.Close()
}

// waitForReady calls the specified endpoint until it gets a 200
// response or until the context is cancelled or the timeout is
// reached.
func waitForReady(
	ctx context.Context,
	timeout time.Duration,
	endpoint string,
) error {
	client := http.Client{}
	startTime := time.Now()
	for {
		req, err := http.NewRequestWithContext(
			ctx,
			http.MethodGet,
			endpoint,
			nil,
		)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("Error making request: %s\n", err.Error())
		} else {
			_ = resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				fmt.Println("Endpoint is ready!")
				return nil
			}
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if time.Since(startTime) >= timeout {
				return fmt.Errorf("timeout reached while waiting for endpoint")
			}
			// wait a little while between checks
			time.Sleep(250 * time.Millisecond)
		}
	}
}
