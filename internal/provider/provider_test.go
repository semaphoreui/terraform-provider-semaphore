package provider

import (
	"fmt"
	"os"
	"terraform-provider-semaphoreui/semaphoreui/client"
	"testing"

	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

var (
	testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
		"semaphoreui": providerserver.NewProtocol6WithError(New("test")()),
	}
)

func mustHaveEnv(t *testing.T, name string) {
	if os.Getenv(name) == "" {
		t.Fatalf("%s environment variable must be set for acceptance tests", name)
	}
}

func testAccPreCheck(t *testing.T) {
	mustHaveEnv(t, "SEMAPHOREUI_API_BASE_URL")
	mustHaveEnv(t, "SEMAPHOREUI_API_TOKEN")
}

var tc *client.SemaphoreUI

func testClient() *client.SemaphoreUI {
	if tc == nil {
		r := httptransport.New(fmt.Sprintf("%s:%s", testHostname(), testPort()), "/api", []string{testProtocol()})
		r.DefaultAuthentication = httptransport.BearerToken(testApiToken())

		tc = client.New(r, strfmt.Default)
	}
	return tc
}

func testHostname() string {
	return os.Getenv("SEMAPHOREUI_HOSTNAME")
}

func testPort() string {
	return os.Getenv("SEMAPHOREUI_PORT")
}

func testProtocol() string {
	return os.Getenv("SEMAPHOREUI_PROTOCOL")
}

func testApiToken() string {
	return os.Getenv("SEMAPHOREUI_API_TOKEN")
}
