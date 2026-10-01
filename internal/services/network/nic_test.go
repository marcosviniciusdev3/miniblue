package network

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/moabukar/miniblue/internal/store"
)

func TestNICOperations(t *testing.T) {
	st := store.New()
	h := NewHandler(st)

	r := chi.NewRouter()
	h.Register(r)

	sub := "00000000-0000-0000-0000-000000000000"
	rg := "test-rg"
	nicName := "test-nic"

	// 1. Create NIC
	nicBody := map[string]interface{}{
		"location": "eastus",
		"properties": map[string]interface{}{
			"ipConfigurations": []interface{}{
				map[string]interface{}{
					"name": "ipconfig1",
					"properties": map[string]interface{}{
						"privateIPAddress":          "10.0.1.10",
						"privateIPAllocationMethod": "Static",
					},
				},
			},
		},
	}
	bodyBytes, _ := json.Marshal(nicBody)

	req := httptest.NewRequest("PUT", "/subscriptions/"+sub+"/resourceGroups/"+rg+"/providers/Microsoft.Network/networkInterfaces/"+nicName, bytes.NewReader(bodyBytes))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201 Created, got %d", rec.Code)
	}

	var created map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &created)
	if created["name"] != nicName {
		t.Errorf("expected name %s, got %v", nicName, created["name"])
	}

	// 2. Get NIC
	req = httptest.NewRequest("GET", "/subscriptions/"+sub+"/resourceGroups/"+rg+"/providers/Microsoft.Network/networkInterfaces/"+nicName, nil)
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200 OK, got %d", rec.Code)
	}

	// 3. List NICs
	req = httptest.NewRequest("GET", "/subscriptions/"+sub+"/resourceGroups/"+rg+"/providers/Microsoft.Network/networkInterfaces", nil)
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200 OK, got %d", rec.Code)
	}

	// 4. Delete NIC
	req = httptest.NewRequest("DELETE", "/subscriptions/"+sub+"/resourceGroups/"+rg+"/providers/Microsoft.Network/networkInterfaces/"+nicName, nil)
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected status 202 Accepted, got %d", rec.Code)
	}
}
