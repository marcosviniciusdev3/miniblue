package vm

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/moabukar/miniblue/internal/store"
)

func TestDiskOperations(t *testing.T) {
	st := store.New()
	h := NewHandler(st)

	r := chi.NewRouter()
	h.Register(r)

	sub := "00000000-0000-0000-0000-000000000000"
	rg := "test-rg"
	diskName := "os-disk-1"

	// 1. Create Disk
	diskBody := map[string]interface{}{
		"location": "eastus",
		"sku": map[string]interface{}{
			"name": "Premium_LRS",
		},
		"properties": map[string]interface{}{
			"diskSizeGB": 256,
			"creationData": map[string]interface{}{
				"createOption": "Empty",
			},
		},
	}
	bodyBytes, _ := json.Marshal(diskBody)

	req := httptest.NewRequest("PUT", "/subscriptions/"+sub+"/resourceGroups/"+rg+"/providers/Microsoft.Compute/disks/"+diskName, bytes.NewReader(bodyBytes))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201 Created, got %d", rec.Code)
	}

	var created map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &created)
	if created["name"] != diskName {
		t.Errorf("expected name %s, got %v", diskName, created["name"])
	}

	props := created["properties"].(map[string]interface{})
	if props["diskSizeGB"].(float64) != 256 {
		t.Errorf("expected diskSizeGB 256, got %v", props["diskSizeGB"])
	}

	// 2. Get Disk
	req = httptest.NewRequest("GET", "/subscriptions/"+sub+"/resourceGroups/"+rg+"/providers/Microsoft.Compute/disks/"+diskName, nil)
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200 OK, got %d", rec.Code)
	}

	// 3. List Disks in RG
	req = httptest.NewRequest("GET", "/subscriptions/"+sub+"/resourceGroups/"+rg+"/providers/Microsoft.Compute/disks", nil)
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200 OK, got %d", rec.Code)
	}

	// 4. Delete Disk
	req = httptest.NewRequest("DELETE", "/subscriptions/"+sub+"/resourceGroups/"+rg+"/providers/Microsoft.Compute/disks/"+diskName, nil)
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected status 202 Accepted, got %d", rec.Code)
	}
}
