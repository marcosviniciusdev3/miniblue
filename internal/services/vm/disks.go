package vm

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/moabukar/miniblue/internal/azerr"
)

func (h *Handler) registerDiskRoutes(r chi.Router) {
	r.Route("/subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/providers/Microsoft.Compute/disks", func(r chi.Router) {
		r.Get("/", h.ListDisks)
		r.Route("/{diskName}", func(r chi.Router) {
			r.Put("/", h.CreateOrUpdateDisk)
			r.Get("/", h.GetDisk)
			r.Delete("/", h.DeleteDisk)
		})
	})
	r.Get("/subscriptions/{subscriptionId}/providers/Microsoft.Compute/disks", h.ListDisksInSubscription)
}

func (h *Handler) diskKey(sub, rg, name string) string {
	return "disk:" + sub + ":" + rg + ":" + name
}

func buildDiskResponse(sub, rg, name string, input map[string]interface{}) map[string]interface{} {
	id := "/subscriptions/" + sub + "/resourceGroups/" + rg + "/providers/Microsoft.Compute/disks/" + name

	location, _ := input["location"].(string)
	if location == "" {
		location = "eastus"
	}

	tags, _ := input["tags"].(map[string]interface{})
	if tags == nil {
		tags = map[string]interface{}{}
	}

	sku, _ := input["sku"].(map[string]interface{})
	if sku == nil {
		sku = map[string]interface{}{
			"name": "Standard_LRS",
			"tier": "Standard",
		}
	}

	props, _ := input["properties"].(map[string]interface{})
	if props == nil {
		props = map[string]interface{}{}
	}

	diskSizeGB, ok := props["diskSizeGB"].(float64)
	if !ok {
		diskSizeGB = 128
	}

	creationData, _ := props["creationData"].(map[string]interface{})
	if creationData == nil {
		creationData = map[string]interface{}{
			"createOption": "Empty",
		}
	}

	diskProps := map[string]interface{}{
		"provisioningState": "Succeeded",
		"uniqueId":          uuid.New().String(),
		"diskSizeGB":        diskSizeGB,
		"diskSizeBytes":     int64(diskSizeGB * 1024 * 1024 * 1024),
		"creationData":      creationData,
		"diskState":         "Unattached",
	}

	if osType, ok := props["osType"]; ok {
		diskProps["osType"] = osType
	}

	return map[string]interface{}{
		"id":         id,
		"name":       name,
		"type":       "Microsoft.Compute/disks",
		"location":   location,
		"tags":       tags,
		"sku":        sku,
		"etag":       "W/\"miniblue\"",
		"properties": diskProps,
	}
}

func (h *Handler) CreateOrUpdateDisk(w http.ResponseWriter, r *http.Request) {
	sub := chi.URLParam(r, "subscriptionId")
	rg := chi.URLParam(r, "resourceGroupName")
	name := chi.URLParam(r, "diskName")

	var input map[string]interface{}
	json.NewDecoder(r.Body).Decode(&input)

	disk := buildDiskResponse(sub, rg, name, input)
	k := h.diskKey(sub, rg, name)
	_, exists := h.store.Get(k)
	h.store.Set(k, disk)

	if exists {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusCreated)
	}
	json.NewEncoder(w).Encode(disk)
}

func (h *Handler) GetDisk(w http.ResponseWriter, r *http.Request) {
	sub := chi.URLParam(r, "subscriptionId")
	rg := chi.URLParam(r, "resourceGroupName")
	name := chi.URLParam(r, "diskName")

	v, ok := h.store.Get(h.diskKey(sub, rg, name))
	if !ok {
		azerr.NotFound(w, "Microsoft.Compute/disks", name)
		return
	}
	json.NewEncoder(w).Encode(v)
}

func (h *Handler) DeleteDisk(w http.ResponseWriter, r *http.Request) {
	sub := chi.URLParam(r, "subscriptionId")
	rg := chi.URLParam(r, "resourceGroupName")
	name := chi.URLParam(r, "diskName")

	if !h.store.Delete(h.diskKey(sub, rg, name)) {
		azerr.NotFound(w, "Microsoft.Compute/disks", name)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

func (h *Handler) ListDisks(w http.ResponseWriter, r *http.Request) {
	sub := chi.URLParam(r, "subscriptionId")
	rg := chi.URLParam(r, "resourceGroupName")
	items := h.store.ListByPrefix("disk:" + sub + ":" + rg + ":")
	json.NewEncoder(w).Encode(map[string]interface{}{"value": items})
}

func (h *Handler) ListDisksInSubscription(w http.ResponseWriter, r *http.Request) {
	sub := chi.URLParam(r, "subscriptionId")
	items := h.store.ListByPrefix("disk:" + sub + ":")
	json.NewEncoder(w).Encode(map[string]interface{}{"value": items})
}
