package client_test

import (
	"context"
	"testing"

	"github.com/numberly/terraform-provider-mica/internal/client"
	"github.com/numberly/terraform-provider-mica/internal/testmock"
	"github.com/numberly/terraform-provider-mica/internal/testmock/handlers"
)

func TestUnit_AuditObjectStorePolicyMember_List(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	store := handlers.RegisterAuditObjectStorePolicyHandlers(ms.Mux)
	c := newTestClient(t, ms.Server)

	store.Seed(&client.AuditObjectStorePolicy{
		ID: "pol-1", Name: "test-policy", Enabled: true, IsLocal: true, PolicyType: "audit",
	})
	store.SeedMember("test-policy", client.AuditObjectStorePolicyMember{
		Member: client.NamedReference{Name: "bucket-a"},
		Policy: client.NamedReference{Name: "test-policy"},
	})

	members, err := c.ListAuditObjectStorePolicyMembers(context.Background(), "test-policy")
	if err != nil {
		t.Fatalf("ListAuditObjectStorePolicyMembers: %v", err)
	}
	if len(members) != 1 {
		t.Fatalf("expected 1 member, got %d", len(members))
	}
	if members[0].Member.Name != "bucket-a" {
		t.Errorf("expected member name bucket-a, got %q", members[0].Member.Name)
	}
}

func TestUnit_AuditObjectStorePolicyMember_List_Empty(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	store := handlers.RegisterAuditObjectStorePolicyHandlers(ms.Mux)
	c := newTestClient(t, ms.Server)

	store.Seed(&client.AuditObjectStorePolicy{
		ID: "pol-2", Name: "empty-policy", Enabled: true, IsLocal: true, PolicyType: "audit",
	})

	members, err := c.ListAuditObjectStorePolicyMembers(context.Background(), "empty-policy")
	if err != nil {
		t.Fatalf("ListAuditObjectStorePolicyMembers: %v", err)
	}
	if len(members) != 0 {
		t.Errorf("expected 0 members, got %d", len(members))
	}
}

func TestUnit_AuditObjectStorePolicyMember_Post(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	store := handlers.RegisterAuditObjectStorePolicyHandlers(ms.Mux)
	c := newTestClient(t, ms.Server)

	store.Seed(&client.AuditObjectStorePolicy{
		ID: "pol-3", Name: "my-policy", Enabled: true, IsLocal: true, PolicyType: "audit",
	})

	member, err := c.PostAuditObjectStorePolicyMember(context.Background(), "my-policy", "my-bucket")
	if err != nil {
		t.Fatalf("PostAuditObjectStorePolicyMember: %v", err)
	}
	if member.Member.Name != "my-bucket" {
		t.Errorf("expected member name my-bucket, got %q", member.Member.Name)
	}
	if member.Policy.Name != "my-policy" {
		t.Errorf("expected policy name my-policy, got %q", member.Policy.Name)
	}
}

func TestUnit_AuditObjectStorePolicyMember_Delete(t *testing.T) {
	ms := testmock.NewMockServer()
	defer ms.Close()
	store := handlers.RegisterAuditObjectStorePolicyHandlers(ms.Mux)
	c := newTestClient(t, ms.Server)

	store.Seed(&client.AuditObjectStorePolicy{
		ID: "pol-4", Name: "del-policy", Enabled: true, IsLocal: true, PolicyType: "audit",
	})
	store.SeedMember("del-policy", client.AuditObjectStorePolicyMember{
		Member: client.NamedReference{Name: "del-bucket"},
		Policy: client.NamedReference{Name: "del-policy"},
	})

	err := c.DeleteAuditObjectStorePolicyMember(context.Background(), "del-policy", "del-bucket")
	if err != nil {
		t.Fatalf("DeleteAuditObjectStorePolicyMember: %v", err)
	}

	members, err := c.ListAuditObjectStorePolicyMembers(context.Background(), "del-policy")
	if err != nil {
		t.Fatalf("ListAuditObjectStorePolicyMembers: %v", err)
	}
	for _, m := range members {
		if m.Member.Name == "del-bucket" {
			t.Error("expected member to be deleted, but it still exists")
		}
	}
}
