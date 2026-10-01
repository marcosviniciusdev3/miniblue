package network

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/moabukar/miniblue/internal/azerr"
)

func (h *Handler) registerNICRoutes(r chi.Router) {
	r.Route("/subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/providers/Microsoft.Network/networkInterfaces", func(r chi.Router) {
		r.Get("/", h.ListNICs)
		r.Route("/{nicName}", func(r chi.Router) {
			r.Put("/", h.CreateOrUpdateNIC)
			r.Get("/", h.GetNIC)
			r.Delete("/", h.DeleteNIC)
		})
	})
	r.Get("/subscriptions/{subscriptionId}/providers/Microsoft.Network/networkInterfaces", h.ListNICsInSubscription)
}

func (h *Handler) nicKey(sub, rg, name string) string {
	return "nic:" + sub + ":" + rg + ":" + name
}

func buildNICResponse(sub, rg, name string, input map[string]interface{}) map[string]interface{} {
	id := "/subscriptions/" + sub + "/resourceGroups/" + rg + "/providers/Microsoft.Network/networkInterfaces/" + name

	location, _ := input["location"].(string)
	if location == "" {
		location = "eastus"
	}

	tags, _ := input["tags"].(map[string]interface{})
	if tags == nil {
		tags = map[string]interface{}{}
	}

	props, _ := input["properties"].(map[string]interface{})
	if props == nil {
		props = map[string]interface{}{}
	}

	ipConfigsRaw, _ := props["ipConfigurations"].([]interface{})
	ipConfigs := make([]interface{}, 0)

	if len(ipConfigsRaw) > 0 {
		for i, ipRaw := range ipConfigsRaw {
			if ipMap, ok := ipRaw.(map[string]interface{}); ok {
				ipName, _ := ipMap["name"].(string)
				if ipName == "" {
					ipName = "ipconfig1"
				}
				ipID := id + "/ipConfigurations/" + ipName

				ipProps, _ := ipMap["properties"].(map[string]interface{})
				if ipProps == nil {
					ipProps = map[string]interface{}{}
				}

				privIP, _ := ipProps["privateIPAddress"].(string)
				if privIP == "" {
					privIP = "10.0.0.4"
				}

				privIPAlloc, _ := ipProps["privateIPAllocationMethod"].(string)
				if privIPAlloc == "" {
					privIPAlloc = "Dynamic"
				}

				constructedIpProps := map[string]interface{}{
					"provisioningState":          "Succeeded",
					"privateIPAddress":          privIP,
					"privateIPAllocationMethod": privIPAlloc,
					"primary":                   i == 0,
				}

				if subnet, ok := ipProps["subnet"]; ok {
					constructedIpProps["subnet"] = subnet
				}
				if pip, ok := ipProps["publicIPAddress"]; ok {
					constructedIpProps["publicIPAddress"] = pip
				}

				ipConfigs = append(ipConfigs, map[string]interface{}{
					"id":         ipID,
					"name":       ipName,
					"etag":       "W/\"miniblue\"",
					"type":       "Microsoft.Network/networkInterfaces/ipConfigurations",
					"properties": constructedIpProps,
				})
			}
		}
	} else {
		// Default IP Config if none provided
		ipConfigs = append(ipConfigs, map[string]interface{}{
			"id":   id + "/ipConfigurations/ipconfig1",
			"name": "ipconfig1",
			"etag": "W/\"miniblue\"",
			"type": "Microsoft.Network/networkInterfaces/ipConfigurations",
			"properties": map[string]interface{}{
				"provisioningState":          "Succeeded",
				"privateIPAddress":          "10.0.0.4",
				"privateIPAllocationMethod": "Dynamic",
				"primary":                   true,
			},
		})
	}

	nicProps := map[string]interface{}{
		"provisioningState":        "Succeeded",
		"resourceGuid":             uuid.New().String(),
		"ipConfigurations":         ipConfigs,
		"dnsSettings":              map[string]interface{}{"dnsServers": []interface{}{}},
		"enableAcceleratedNetworking": false,
		"enableIPForwarding":       false,
	}

	if nsg, ok := props["networkSecurityGroup"]; ok {
		nicProps["networkSecurityGroup"] = nsg
	}

	return map[string]interface{}{
		"id":         id,
		"name":       name,
		"type":       "Microsoft.Network/networkInterfaces",
		"location":   location,
		"tags":       tags,
		"etag":       "W/\"miniblue\"",
		"properties": nicProps,
	}
}

func (h *Handler) CreateOrUpdateNIC(w http.ResponseWriter, r *http.Request) {
	sub := chi.URLParam(r, "subscriptionId")
	rg := chi.URLParam(r, "resourceGroupName")
	name := chi.URLParam(r, "nicName")

	var input map[string]interface{}
	json.NewDecoder(r.Body).Decode(&input)

	nic := buildNICResponse(sub, rg, name, input)
	k := h.nicKey(sub, rg, name)
	_, exists := h.store.Get(k)
	h.store.Set(k, nic)

	if exists {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusCreated)
	}
	json.NewEncoder(w).Encode(nic)
}

func (h *Handler) GetNIC(w http.ResponseWriter, r *http.Request) {
	sub := chi.URLParam(r, "subscriptionId")
	rg := chi.URLParam(r, "resourceGroupName")
	name := chi.URLParam(r, "nicName")

	v, ok := h.store.Get(h.nicKey(sub, rg, name))
	if !ok {
		azerr.NotFound(w, "Microsoft.Network/networkInterfaces", name)
		return
	}
	json.NewEncoder(w).Encode(v)
}

func (h *Handler) DeleteNIC(w http.ResponseWriter, r *http.Request) {
	sub := chi.URLParam(r, "subscriptionId")
	rg := chi.URLParam(r, "resourceGroupName")
	name := chi.URLParam(r, "nicName")

	if !h.store.Delete(h.nicKey(sub, rg, name)) {
		azerr.NotFound(w, "Microsoft.Network/networkInterfaces", name)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

func (h *Handler) ListNICs(w http.ResponseWriter, r *http.Request) {
	sub := chi.URLParam(r, "subscriptionId")
	rg := chi.URLParam(r, "resourceGroupName")
	items := h.store.ListByPrefix("nic:" + sub + ":" + rg + ":")
	json.NewEncoder(w).Encode(map[string]interface{}{"value": items})
}

func (h *Handler) ListNICsInSubscription(w http.ResponseWriter, r *http.Request) {
	sub := chi.URLParam(r, "subscriptionId")
	items := h.store.ListByPrefix("nic:" + sub + ":")
	json.NewEncoder(w).Encode(map[string]interface{}{"value": items})
}
