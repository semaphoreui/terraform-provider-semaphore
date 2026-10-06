package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	apiclient "terraform-provider-semaphoreui/semaphoreui/client"
)

func TestUserResourceUpdatePasswordFailurePreservesVersion(t *testing.T) {
	ctx := context.Background()
	readAfterFailure := false
	passwordRequests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		requestPath := strings.TrimSuffix(req.URL.Path, "/")
		switch {
		case req.Method == http.MethodPut && requestPath == "/api/users/1":
			w.WriteHeader(http.StatusNoContent)
		case req.Method == http.MethodPost && requestPath == "/api/users/1/password":
			passwordRequests++
			w.WriteHeader(http.StatusInternalServerError)
		case req.Method == http.MethodGet && requestPath == "/api/users/1":
			readAfterFailure = true
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"id":1,"username":"test","name":"Test","email":"test@example.com","created":"2026-01-01","admin":false,"alert":false,"external":false}`))
		default:
			t.Errorf("unexpected request: %s %s", req.Method, req.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()
	endpoint, err := url.Parse(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	r := &userResource{client: apiclient.New(httptransport.New(endpoint.Host, "/api", []string{"http"}), strfmt.Default)}
	schema := userSchema().GetResource(ctx)
	model := UserModel{
		ID: types.Int64Value(1), Created: types.StringValue("2026-01-01"),
		Username: types.StringValue("test"), Name: types.StringValue("Test"),
		Email: types.StringValue("test@example.com"), Admin: types.BoolValue(false),
		External: types.BoolValue(false), Alert: types.BoolValue(false),
		Password: types.StringNull(), PasswordWO: types.StringNull(),
		PasswordWOVersion: types.Int64Value(1),
	}
	previous := tfsdk.State{Schema: schema}
	if diags := previous.Set(ctx, model); diags.HasError() {
		t.Fatal(diags)
	}
	model.PasswordWOVersion = types.Int64Value(2)
	planned := tfsdk.State{Schema: schema}
	if diags := planned.Set(ctx, model); diags.HasError() {
		t.Fatal(diags)
	}
	model.PasswordWO = types.StringValue("rotated-test-password")
	configured := tfsdk.State{Schema: schema}
	if diags := configured.Set(ctx, model); diags.HasError() {
		t.Fatal(diags)
	}
	response := resource.UpdateResponse{State: previous}
	r.Update(ctx, resource.UpdateRequest{
		State:  previous,
		Plan:   tfsdk.Plan{Schema: schema, Raw: planned.Raw},
		Config: tfsdk.Config{Schema: schema, Raw: configured.Raw},
	}, &response)
	if !response.Diagnostics.HasError() {
		t.Fatal("expected password rotation error")
	}
	if passwordRequests != 1 {
		t.Fatalf("expected one password request, got %d", passwordRequests)
	}
	var actualVersion types.Int64
	if diags := response.State.GetAttribute(ctx, path.Root("password_wo_version"), &actualVersion); diags.HasError() {
		t.Fatal(diags)
	}
	if !actualVersion.Equal(types.Int64Value(1)) {
		t.Fatalf("failed rotation advanced version to %s", actualVersion)
	}
	if readAfterFailure {
		t.Fatal("user was read after failed password rotation")
	}
}

func TestUserPasswordVersionSchema(t *testing.T) {
	ctx := context.Background()
	resourceAttribute := userSchema().GetResource(ctx).Attributes["password_wo_version"]
	if !resourceAttribute.IsOptional() || resourceAttribute.IsComputed() {
		t.Fatal("resource password version must be optional and not computed")
	}
	dataSourceAttribute := userSchema().GetDataSource(ctx).Attributes["password_wo_version"]
	if dataSourceAttribute.IsOptional() || !dataSourceAttribute.IsComputed() {
		t.Fatal("data source password version must be computed-only")
	}
}

func TestUserResourceUpdateRemovingPasswordInputs(t *testing.T) {
	for _, test := range []struct {
		name     string
		password types.String
		version  types.Int64
	}{
		{name: "write-only", password: types.StringNull(), version: types.Int64Value(1)},
		{name: "ordinary", password: types.StringValue("previous-password"), version: types.Int64Null()},
	} {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			passwordRequests := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				switch requestPath := strings.TrimSuffix(req.URL.Path, "/"); {
				case req.Method == http.MethodPut && requestPath == "/api/users/1":
					w.WriteHeader(http.StatusNoContent)
				case req.Method == http.MethodGet && requestPath == "/api/users/1":
					w.Header().Set("Content-Type", "application/json")
					_, _ = w.Write([]byte(`{"id":1,"username":"test","name":"Test","email":"test@example.com","created":"2026-01-01","admin":false,"alert":false,"external":false}`))
				case req.Method == http.MethodPost && requestPath == "/api/users/1/password":
					passwordRequests++
					w.WriteHeader(http.StatusInternalServerError)
				default:
					t.Errorf("unexpected request: %s %s", req.Method, req.URL.Path)
					w.WriteHeader(http.StatusNotFound)
				}
			}))
			defer server.Close()
			endpoint, err := url.Parse(server.URL)
			if err != nil {
				t.Fatal(err)
			}
			r := &userResource{client: apiclient.New(httptransport.New(endpoint.Host, "/api", []string{"http"}), strfmt.Default)}
			schema := userSchema().GetResource(ctx)
			model := UserModel{
				ID: types.Int64Value(1), Created: types.StringValue("2026-01-01"),
				Username: types.StringValue("test"), Name: types.StringValue("Test"),
				Email: types.StringValue("test@example.com"), Admin: types.BoolValue(false),
				External: types.BoolValue(false), Alert: types.BoolValue(false),
				Password: test.password, PasswordWO: types.StringNull(), PasswordWOVersion: test.version,
			}
			previous := tfsdk.State{Schema: schema}
			if diags := previous.Set(ctx, model); diags.HasError() {
				t.Fatal(diags)
			}
			model.Password = types.StringNull()
			model.PasswordWOVersion = types.Int64Null()
			planned := tfsdk.State{Schema: schema}
			if diags := planned.Set(ctx, model); diags.HasError() {
				t.Fatal(diags)
			}
			response := resource.UpdateResponse{State: previous}
			r.Update(ctx, resource.UpdateRequest{
				State: previous, Plan: tfsdk.Plan{Schema: schema, Raw: planned.Raw},
				Config: tfsdk.Config{Schema: schema, Raw: planned.Raw},
			}, &response)
			if response.Diagnostics.HasError() {
				t.Fatalf("removing password inputs failed: %v", response.Diagnostics)
			}
			if passwordRequests != 0 {
				t.Fatalf("removing password inputs sent %d password requests", passwordRequests)
			}
			var actualVersion types.Int64
			if diags := response.State.GetAttribute(ctx, path.Root("password_wo_version"), &actualVersion); diags.HasError() {
				t.Fatal(diags)
			}
			if !actualVersion.IsNull() {
				t.Fatalf("removed password version remains in state: %s", actualVersion)
			}
		})
	}
}
