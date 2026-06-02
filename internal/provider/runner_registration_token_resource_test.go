package provider

import (
	"fmt"
	"terraform-provider-semaphoreui/semaphoreui/client/runner"
	"terraform-provider-semaphoreui/semaphoreui/models"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// testAccPreCheckRunnerRegistrationToken skips the test when the SemaphoreUI
// server under test does not expose the runner registration-token endpoint
// (it was added in a later release than some versions covered by CI).
func testAccPreCheckRunnerRegistrationToken(t *testing.T) {
	testAccPreCheck(t)

	client := testClient()
	created, err := client.Runner.PostRunners(&runner.PostRunnersParams{
		Runner: &models.RunnerRequest{Name: "regtoken-precheck-" + acctest.RandString(6)},
	}, nil)
	if err != nil {
		// Could not probe; let the test run and surface any real error.
		return
	}
	runnerID := created.Payload.ID
	defer func() {
		_, _ = client.Runner.DeleteRunnersRunnerID(&runner.DeleteRunnersRunnerIDParams{
			RunnerID: runnerID,
		}, nil)
	}()

	if _, err := client.Runner.PostRunnersRunnerIDRegistrationToken(&runner.PostRunnersRunnerIDRegistrationTokenParams{
		RunnerID: runnerID,
	}, nil); err != nil {
		t.Skipf("skipping: SemaphoreUI runner registration-token endpoint unavailable: %s", err.Error())
	}
}

func testAccCaptureAttr(resourceName, attr string, dest *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}
		*dest = rs.Primary.Attributes[attr]
		return nil
	}
}

func testAccRunnerRegistrationTokenConfig(nameSuffix, rotation string) string {
	return fmt.Sprintf(`
resource "semaphoreui_runner" "test" {
  name   = "Test %[1]s"
  active = false
}
resource "semaphoreui_runner_registration_token" "test" {
  runner_id = semaphoreui_runner.test.id
  keepers = {
    rotation = "%[2]s"
  }
}`, nameSuffix, rotation)
}

func TestAcc_RunnerRegistrationTokenResource_basic(t *testing.T) {
	nameSuffix := acctest.RandString(8)
	var token1 string
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheckRunnerRegistrationToken(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccRunnerRegistrationTokenConfig(nameSuffix, "1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("semaphoreui_runner_registration_token.test", "registration_token"),
					resource.TestCheckResourceAttrSet("semaphoreui_runner_registration_token.test", "runner_id"),
					resource.TestCheckResourceAttrSet("semaphoreui_runner_registration_token.test", "id"),
					resource.TestCheckResourceAttr("semaphoreui_runner_registration_token.test", "keepers.rotation", "1"),
					testAccCaptureAttr("semaphoreui_runner_registration_token.test", "registration_token", &token1),
				),
			},
			// Rotation: changing keepers forces a brand new token.
			{
				Config: testAccRunnerRegistrationTokenConfig(nameSuffix, "2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("semaphoreui_runner_registration_token.test", "keepers.rotation", "2"),
					resource.TestCheckResourceAttrWith("semaphoreui_runner_registration_token.test", "registration_token", func(v string) error {
						if v == "" {
							return fmt.Errorf("registration_token is empty")
						}
						if v == token1 {
							return fmt.Errorf("registration_token was not rotated")
						}
						return nil
					}),
				),
			},
		},
	})
}
