package client_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/numberly/opentofu-provider-flashblade/internal/client"
)

// TestUnit_QosPolicyPost_JSONEncoding verifies that QosPolicyPost marshalling:
//   - omits the "name" key (name goes via ?names= query param, json:"-")
//   - preserves MaxTotalBytesPerSec=0 (unlimited) when pointer is non-nil
//   - omits MaxTotal* fields when pointer is nil
func TestUnit_QosPolicyPost_JSONEncoding(t *testing.T) {
	// Case 1: name must NOT appear, zero pointer values must serialize.
	zero := int64(0)
	body := client.QosPolicyPost{
		Name:                "should-not-appear",
		MaxTotalBytesPerSec: &zero,
		MaxTotalOpsPerSec:   &zero,
	}
	buf, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	s := string(buf)
	if strings.Contains(s, `"name"`) {
		t.Errorf("expected no \"name\" key in QosPolicyPost JSON, got: %s", s)
	}
	if !strings.Contains(s, `"max_total_bytes_per_sec":0`) {
		t.Errorf("expected max_total_bytes_per_sec:0 preserved, got: %s", s)
	}
	if !strings.Contains(s, `"max_total_ops_per_sec":0`) {
		t.Errorf("expected max_total_ops_per_sec:0 preserved, got: %s", s)
	}

	// Case 2: nil pointers must omit fields entirely.
	body2 := client.QosPolicyPost{Name: "x"}
	buf2, err := json.Marshal(body2)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	s2 := string(buf2)
	if strings.Contains(s2, "max_total_bytes_per_sec") {
		t.Errorf("expected max_total_bytes_per_sec omitted when nil, got: %s", s2)
	}
	if strings.Contains(s2, "max_total_ops_per_sec") {
		t.Errorf("expected max_total_ops_per_sec omitted when nil, got: %s", s2)
	}
}

func TestUnit_QosPolicy_Get(t *testing.T) {
	expected := client.QosPolicy{
		ID:                  "qos-id-001",
		Name:                "high-throughput",
		Enabled:             true,
		IsLocal:             true,
		MaxTotalBytesPerSec: 1073741824, // 1 GiB/s
		MaxTotalOpsPerSec:   10000,
		PolicyType:          "qos",
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/api/2.22/qos-policies":
			name := r.URL.Query().Get("names")
			if name != "high-throughput" {
				writeJSON(w, http.StatusOK, listResponse([]client.QosPolicy{}))
				return
			}
			writeJSON(w, http.StatusOK, listResponse([]client.QosPolicy{expected}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	got, err := c.GetQosPolicy(context.Background(), "high-throughput")
	if err != nil {
		t.Fatalf("GetQosPolicy: %v", err)
	}
	if got.ID != expected.ID {
		t.Errorf("expected ID %q, got %q", expected.ID, got.ID)
	}
	if got.MaxTotalBytesPerSec != 1073741824 {
		t.Errorf("expected MaxTotalBytesPerSec 1073741824, got %d", got.MaxTotalBytesPerSec)
	}
}

func TestUnit_QosPolicy_Get_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/api/2.22/qos-policies":
			writeJSON(w, http.StatusOK, listResponse([]client.QosPolicy{}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	_, err := c.GetQosPolicy(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for not found, got nil")
	}
	if !client.IsNotFound(err) {
		t.Errorf("expected IsNotFound true, got false; err: %v", err)
	}
}

func TestUnit_QosPolicy_Post(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodPost && r.URL.Path == "/api/2.22/qos-policies":
			name := r.URL.Query().Get("names")
			if name == "" {
				http.Error(w, "names query param required", http.StatusBadRequest)
				return
			}
			var body client.QosPolicyPost
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, "bad body", http.StatusBadRequest)
				return
			}
			enabled := true
			if body.Enabled != nil {
				enabled = *body.Enabled
			}
			var maxBytes, maxOps int64
			if body.MaxTotalBytesPerSec != nil {
				maxBytes = *body.MaxTotalBytesPerSec
			}
			if body.MaxTotalOpsPerSec != nil {
				maxOps = *body.MaxTotalOpsPerSec
			}
			result := client.QosPolicy{
				ID:                  "qos-id-002",
				Name:                name,
				Enabled:             enabled,
				MaxTotalBytesPerSec: maxBytes,
				MaxTotalOpsPerSec:   maxOps,
			}
			writeJSON(w, http.StatusOK, listResponse([]client.QosPolicy{result}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	enabled := true
	maxBytes := int64(536870912) // 512 MiB/s
	maxOps := int64(5000)
	got, err := c.PostQosPolicy(context.Background(), "new-policy", client.QosPolicyPost{
		Enabled:             &enabled,
		MaxTotalBytesPerSec: &maxBytes,
		MaxTotalOpsPerSec:   &maxOps,
	})
	if err != nil {
		t.Fatalf("PostQosPolicy: %v", err)
	}
	if got.Name != "new-policy" {
		t.Errorf("expected Name new-policy, got %q", got.Name)
	}
	if got.MaxTotalBytesPerSec != 536870912 {
		t.Errorf("expected MaxTotalBytesPerSec 536870912, got %d", got.MaxTotalBytesPerSec)
	}
}

func TestUnit_QosPolicy_Patch(t *testing.T) {
	var gotBody client.QosPolicyPatch

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodPatch && r.URL.Path == "/api/2.22/qos-policies":
			name := r.URL.Query().Get("names")
			if name != "high-throughput" {
				http.Error(w, "unexpected names param", http.StatusBadRequest)
				return
			}
			if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
				http.Error(w, "bad body", http.StatusBadRequest)
				return
			}
			newMax := int64(2147483648)
			if gotBody.MaxTotalBytesPerSec != nil {
				newMax = *gotBody.MaxTotalBytesPerSec
			}
			result := client.QosPolicy{
				ID:                  "qos-id-001",
				Name:                name,
				Enabled:             true,
				MaxTotalBytesPerSec: newMax,
			}
			writeJSON(w, http.StatusOK, listResponse([]client.QosPolicy{result}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	newMax := int64(2147483648) // 2 GiB/s
	got, err := c.PatchQosPolicy(context.Background(), "high-throughput", client.QosPolicyPatch{
		MaxTotalBytesPerSec: &newMax,
	})
	if err != nil {
		t.Fatalf("PatchQosPolicy: %v", err)
	}
	if got.MaxTotalBytesPerSec != 2147483648 {
		t.Errorf("expected MaxTotalBytesPerSec 2147483648, got %d", got.MaxTotalBytesPerSec)
	}
	// PATCH semantics: MaxTotalOpsPerSec should be absent
	if gotBody.MaxTotalOpsPerSec != nil {
		t.Errorf("expected MaxTotalOpsPerSec absent in PATCH body")
	}
}

func TestUnit_QosPolicy_Delete(t *testing.T) {
	var deleteCalled bool

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodDelete && r.URL.Path == "/api/2.22/qos-policies":
			name := r.URL.Query().Get("names")
			if name != "high-throughput" {
				http.Error(w, "unexpected names param", http.StatusBadRequest)
				return
			}
			deleteCalled = true
			w.WriteHeader(http.StatusOK)
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	if err := c.DeleteQosPolicy(context.Background(), "high-throughput"); err != nil {
		t.Fatalf("DeleteQosPolicy: %v", err)
	}
	if !deleteCalled {
		t.Error("expected DELETE to be called")
	}
}

func TestUnit_QosPolicyMember_List(t *testing.T) {
	members := []client.QosPolicyMember{
		{
			Member: client.NamedReference{Name: "fs-001", ID: "fs-id-001"},
			Policy: client.NamedReference{Name: "high-throughput", ID: "qos-id-001"},
		},
		{
			Member: client.NamedReference{Name: "fs-002", ID: "fs-id-002"},
			Policy: client.NamedReference{Name: "high-throughput", ID: "qos-id-001"},
		},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/api/2.22/qos-policies/members":
			policyName := r.URL.Query().Get("policy_names")
			if policyName != "high-throughput" {
				writeJSON(w, http.StatusOK, map[string]any{"items": []client.QosPolicyMember{}})
				return
			}
			writeJSON(w, http.StatusOK, map[string]any{
				"items":            members,
				"total_item_count": 2,
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	got, err := c.ListQosPolicyMembers(context.Background(), "high-throughput")
	if err != nil {
		t.Fatalf("ListQosPolicyMembers: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("expected 2 members, got %d", len(got))
	}
	if got[0].Member.Name != "fs-001" {
		t.Errorf("expected first member fs-001, got %q", got[0].Member.Name)
	}
}

func TestUnit_QosPolicyMember_Post(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodPost && r.URL.Path == "/api/2.22/qos-policies/members":
			policyName := r.URL.Query().Get("policy_names")
			memberName := r.URL.Query().Get("member_names")
			memberType := r.URL.Query().Get("member_types")
			if policyName == "" || memberName == "" || memberType == "" {
				http.Error(w, "policy_names, member_names, and member_types required", http.StatusBadRequest)
				return
			}
			result := client.QosPolicyMember{
				Member: client.NamedReference{Name: memberName},
				Policy: client.NamedReference{Name: policyName},
			}
			writeJSON(w, http.StatusOK, listResponse([]client.QosPolicyMember{result}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	got, err := c.PostQosPolicyMember(context.Background(), "high-throughput", "fs-003", "file-systems")
	if err != nil {
		t.Fatalf("PostQosPolicyMember: %v", err)
	}
	if got.Member.Name != "fs-003" {
		t.Errorf("expected Member.Name fs-003, got %q", got.Member.Name)
	}
	if got.Policy.Name != "high-throughput" {
		t.Errorf("expected Policy.Name high-throughput, got %q", got.Policy.Name)
	}
}

func TestUnit_QosPolicyMember_Delete(t *testing.T) {
	var deleteCalled bool

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/login":
			w.Header().Set("x-auth-token", "tok")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodDelete && r.URL.Path == "/api/2.22/qos-policies/members":
			policyName := r.URL.Query().Get("policy_names")
			memberName := r.URL.Query().Get("member_names")
			if policyName != "high-throughput" || memberName != "fs-003" {
				http.Error(w, "unexpected params", http.StatusBadRequest)
				return
			}
			deleteCalled = true
			w.WriteHeader(http.StatusOK)
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	if err := c.DeleteQosPolicyMember(context.Background(), "high-throughput", "fs-003"); err != nil {
		t.Fatalf("DeleteQosPolicyMember: %v", err)
	}
	if !deleteCalled {
		t.Error("expected DELETE to be called")
	}
}
