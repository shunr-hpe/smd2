/*
 * Copyright © 2026 OpenCHAMI a Series of LF Projects, LLC
 *
 * SPDX-License-Identifier: MIT
 *
 * Test helpers for the csmschema package.  The server is started with
 * --schema pointing at schemas/csm so that the CSM JSON-schema overrides
 * (enum constraints, MAC-address pattern, etc.) are active for the duration
 * of every test in this package.
 */

package csmschema

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

// testServer is the shared httptest.Server wrapping a reverse proxy to the
// running server.  All tests use testServer.URL as the base URL.
var testServer *httptest.Server

// TestMain sets up and tears down the server process for all tests in this
// package.  The server is started with --schema pointing at schemas/csm so
// that the CSM-specific schema constraints are active.
func TestMain(m *testing.M) {
	// resttests/csmschema/ is two levels below the project root.
	projectRoot := "../.."

	absProjectRoot, err := filepath.Abs(projectRoot)
	if err != nil {
		fmt.Printf("Failed to resolve project root: %v\n", err)
		os.Exit(1)
	}

	// ── 1. Clear and recreate the data directory ─────────────────────────────
	absDataDir := filepath.Join(absProjectRoot, "data-csmschema-resttests")
	if err := os.RemoveAll(absDataDir); err != nil {
		fmt.Printf("Failed to remove data directory: %v\n", err)
		os.Exit(1)
	}
	if err := os.MkdirAll(absDataDir, 0o755); err != nil {
		fmt.Printf("Failed to create data directory: %v\n", err)
		os.Exit(1)
	}

	// ── 2. Build the server binary ────────────────────────────────────────────
	binaryName := "inventory-service-csmschema-test"
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}
	absBinaryPath := filepath.Join(absProjectRoot, "bin", binaryName)

	buildCmd := exec.Command("go", "build", "-o", absBinaryPath, "./cmd/server")
	buildCmd.Dir = absProjectRoot
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

	// ── 4. Resolve the CSM schema directory ──────────────────────────────────
	absCsmSchemaDir := filepath.Join(absProjectRoot, "schemas", "csm")

	// ── 5. Start the server subprocess ───────────────────────────────────────
	dbURL := fmt.Sprintf("file:%s?_fk=1", filepath.Join(absDataDir, "data.db"))

	serverCmd := exec.Command(
		absBinaryPath,
		"serve",
		"--port", fmt.Sprintf("%d", port),
		"--host", "127.0.0.1",
		"--database-url", dbURL,
		"--schema", absCsmSchemaDir,
	)
	serverCmd.Dir = absProjectRoot
	serverCmd.Stdout = os.Stdout
	serverCmd.Stderr = os.Stderr

	if err := serverCmd.Start(); err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Server started (PID %d) with CSM schemas from %s\n", serverCmd.Process.Pid, absCsmSchemaDir)

	// ── 6. Wait for the server to become ready ────────────────────────────────
	rawServerURL := fmt.Sprintf("http://127.0.0.1:%d", port)
	ready := false
	for i := 0; i < 30; i++ {
		resp, err := http.Get(rawServerURL + "/health")
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			ready = true
			break
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(200 * time.Millisecond)
	}
	if !ready {
		fmt.Println("Server did not become ready in time")
		serverCmd.Process.Kill() //nolint:errcheck
		os.Exit(1)
	}
	fmt.Printf("Server ready at %s\n", rawServerURL)

	// ── 7. Wrap the subprocess with httptest.NewServer via a reverse proxy ────
	target, _ := url.Parse(rawServerURL)
	proxy := httputil.NewSingleHostReverseProxy(target)
	testServer = httptest.NewServer(proxy)
	defer testServer.Close()

	// ── 8. Run all tests ──────────────────────────────────────────────────────
	code := m.Run()
	serverCmd.Process.Kill() //nolint:errcheck
	os.Exit(code)
}

// ─── Request helpers ─────────────────────────────────────────────────────────

// doRequest encodes body as JSON, fires the request, and returns the response.
// The caller is responsible for closing resp.Body.
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

// assertSchemaError posts body to path and fails the test if the server returns
// a 2xx status, which would indicate that schema validation was not enforced.
func assertSchemaError(t *testing.T, path string, body interface{}) {
	t.Helper()
	resp := doRequest(t, http.MethodPost, path, body)
	defer resp.Body.Close()
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		t.Errorf("expected 4xx for schema-violating request to %s, got HTTP %d", path, resp.StatusCode)
	}
}

// assertSchemaOK posts body to path and fails the test if the server does NOT
// return a 2xx status.
func assertSchemaOK(t *testing.T, path string, body interface{}) {
	t.Helper()
	resp := doRequest(t, http.MethodPost, path, body)
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var buf bytes.Buffer
		buf.ReadFrom(resp.Body)
		t.Errorf("expected 2xx for valid request to %s, got HTTP %d; body: %s", path, resp.StatusCode, buf.String())
	}
}
