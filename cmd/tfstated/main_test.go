package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"testing"
	"time"

	"git.adyxax.org/adyxax/tfstated/pkg/database"
)

var baseURI = url.URL{
	Host:   "127.0.0.1:8081",
	Path:   "/",
	Scheme: "http",
}

func TestMain(m *testing.M) {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	config := Config{
		Host: "127.0.0.1",
		Port: "8081",
	}
	_ = os.Remove("./test.db")
	db, err := database.NewDB(ctx, "./test.db")
	if err != nil {
		fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(1)
	}
	getenv := func(key string) string {
		switch key {
		case "DATA_ENCRYPTION_KEY":
			return "hP3ZSCnY3LMgfTQjwTaGrhKwdA0yXMXIfv67OJnntqM="
		default:
			return ""
		}
	}

	go run(
		ctx,
		&config,
		db,
		getenv,
		os.Stderr,
	)
	err = waitForReady(ctx, 5*time.Second, "http://127.0.0.1:8081/healthz")
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

func runHTTPRequest(method string, uriRef *url.URL, body io.Reader, testFunc func(*http.Response, error)) {
	uri := baseURI.ResolveReference(uriRef)
	client := http.Client{}
	req, err := http.NewRequest(method, uri.String(), body)
	if err != nil {
		testFunc(nil, fmt.Errorf("failed to create request: %w", err))
		return
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
			continue
		}
		if resp.StatusCode == http.StatusOK {
			fmt.Println("Endpoint is ready!")
			resp.Body.Close()
			return nil
		}
		resp.Body.Close()

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
