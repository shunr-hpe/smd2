/*
 * Copyright © 2026 OpenCHAMI a Series of LF Projects, LLC
 *
 * SPDX-License-Identifier: MIT
 */

package resttests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

// testServer is the shared httptest.Server wrapping a reverse proxy to the running server.
// All tests use testServer.URL as the base URL.
var testServer *httptest.Server

// TestMain sets up and tears down the server process for all tests in the package.
// It clears the data-resttests directory, builds the server binary, starts the server,
// wraps it with httptest.NewServer via a reverse proxy, then runs all tests.
func TestMain(m *testing.M) {
	// Locate project root relative to this test file (resttests/ is one level below root)
	projectRoot := ".."

	// Resolve absolute path for the data directory so SQLite can locate it reliably
	absDataDir, err := filepath.Abs(filepath.Join(projectRoot, "data-resttests"))
	if err != nil {
		fmt.Printf("Failed to resolve data directory path: %v\n", err)
		os.Exit(1)
	}

	// ── 1. Clear and recreate the data directory ─────────────────────────────
	if err := os.RemoveAll(absDataDir); err != nil {
		fmt.Printf("Failed to remove data directory: %v\n", err)
		os.Exit(1)
	}
	if err := os.MkdirAll(absDataDir, 0o755); err != nil {
		fmt.Printf("Failed to create data directory: %v\n", err)
		os.Exit(1)
	}

	// ── 2. Build the server binary ────────────────────────────────────────────
	binaryName := "inventory-service-test"
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}
	absBinaryPath, err := filepath.Abs(filepath.Join(projectRoot, "bin", binaryName))
	if err != nil {
		fmt.Printf("Failed to resolve binary path: %v\n", err)
		os.Exit(1)
	}

	buildCmd := exec.Command("go", "build", "-o", absBinaryPath, "./cmd/server")
	buildCmd.Dir, err = filepath.Abs(projectRoot)
	if err != nil {
		fmt.Printf("Failed to resolve project root: %v\n", err)
		os.Exit(1)
	}
	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr
	if err := buildCmd.Run(); err != nil {
		fmt.Printf("Failed to build server binary: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Server binary built: %s\n", absBinaryPath)

	// ── 3. Find a free TCP port ───────────────────────────────────────────────
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		fmt.Printf("Failed to find a free port: %v\n", err)
		os.Exit(1)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()

	// ── 4. Start the server subprocess ───────────────────────────────────────
	// SQLite database stored in the data-resttests directory
	// dbURL := fmt.Sprintf("file:%s?cache=shared&_fk=1", filepath.Join(absDataDir, "data.db"))
	dbURL := fmt.Sprintf("file:%s?_fk=1", filepath.Join(absDataDir, "data.db"))

	serverCmd := exec.Command(
		absBinaryPath,
		"serve",
		"--port", fmt.Sprintf("%d", port),
		"--host", "127.0.0.1",
		"--database-url", dbURL,
	)
	serverCmd.Dir, _ = filepath.Abs(projectRoot)
	serverCmd.Stdout = os.Stdout
	serverCmd.Stderr = os.Stderr

	if err := serverCmd.Start(); err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
		os.Exit(1)
	}
	defer serverCmd.Process.Kill() //nolint:errcheck

	// ── 5. Wait for the server to become ready ────────────────────────────────
	rawServerURL := fmt.Sprintf("http://127.0.0.1:%d", port)
	ready := false
	for i := 0; i < 50; i++ {
		resp, err := http.Get(rawServerURL + "/health") //nolint:noctx
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				ready = true
				break
			}
		}
		time.Sleep(200 * time.Millisecond)
	}
	if !ready {
		fmt.Println("Server did not become ready in time")
		serverCmd.Process.Kill() //nolint:errcheck
		os.Exit(1)
	}
	fmt.Printf("Server ready at %s\n", rawServerURL)

	// ── 6. Wrap the subprocess with httptest.NewServer via a reverse proxy ────
	// This lets tests use a standard *httptest.Server and http.Client pattern.
	target, _ := url.Parse(rawServerURL)
	proxy := httputil.NewSingleHostReverseProxy(target)
	testServer = httptest.NewServer(proxy)
	defer testServer.Close()

	// ── 7. Run all tests ──────────────────────────────────────────────────────
	code := m.Run()
	serverCmd.Process.Kill() //nolint:errcheck
	os.Exit(code)
}

// ─── Request helpers ─────────────────────────────────────────────────────────

// doRequest is a convenience wrapper that encodes body as JSON, fires the
// request, and returns the response. The caller is responsible for closing
// resp.Body.
func doRequest(t *testing.T, method, path string, body interface{}) *http.Response {
	t.Helper()

	var reqBody []byte
	if body != nil {
		var err error
		reqBody, err = json.Marshal(body)
		if err != nil {
			t.Fatalf("failed to marshal request body: %v", err)
		}
	}

	req, err := http.NewRequest(method, testServer.URL+path, bytes.NewReader(reqBody))
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("failed to execute request %s %s: %v", method, path, err)
	}
	return resp
}

// decodeJSON decodes the response body into dst and closes the body.
func decodeJSON(t *testing.T, resp *http.Response, dst interface{}) {
	t.Helper()
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(dst); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}
}

// requireStatus fails the test if the response status does not match expected.
func requireStatus(t *testing.T, resp *http.Response, expected int) {
	t.Helper()
	if resp.StatusCode != expected {
		defer resp.Body.Close()
		var buf bytes.Buffer
		buf.ReadFrom(resp.Body)
		t.Fatalf("expected HTTP %d, got %d; body: %s", expected, resp.StatusCode, buf.String())
	}
}
