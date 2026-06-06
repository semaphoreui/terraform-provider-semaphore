package provider

import (
	"fmt"
	"strconv"
	"terraform-provider-semaphoreui/semaphoreui/client/runner"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func testAccRunnerExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		if rs.Primary.Attributes["id"] == "" {
			return fmt.Errorf("no ID is set")
		}

		id, _ := strconv.ParseInt(rs.Primary.Attributes["id"], 10, 64)

		_, err := testClient().Runner.GetRunnersRunnerID(&runner.GetRunnersRunnerIDParams{
			RunnerID: id,
		}, nil)
		if err != nil {
			return fmt.Errorf("error reading runner: %s", err.Error())
		}

		return nil
	}
}

func testAccRunnerConfig(nameSuffix string, maxParallelTasks int, active bool, isDefault bool, tags string) string {
	return fmt.Sprintf(`
resource "semaphoreui_runner" "test" {
  name               = "Test %[1]s"
  max_parallel_tasks = %[2]d
  active             = %[3]t
  is_default         = %[4]t
  tags               = %[5]s
}`, nameSuffix, maxParallelTasks, active, isDefault, tags)
}

func testAccCheckRunnerDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "semaphoreui_runner" {
			continue
		}

		id, _ := strconv.ParseInt(rs.Primary.Attributes["id"], 10, 64)
		_, err := testClient().Runner.GetRunnersRunnerID(&runner.GetRunnersRunnerIDParams{
			RunnerID: id,
		}, nil)
		if err == nil {
			return fmt.Errorf("runner %d still exists", id)
		}
	}
	return nil
}

// testAccDeleteRunnerOutOfBand deletes the runner directly via the API to
// simulate a deletion performed outside Terraform (e.g. from the web UI).
func testAccDeleteRunnerOutOfBand(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}
		id, _ := strconv.ParseInt(rs.Primary.Attributes["id"], 10, 64)
		_, err := testClient().Runner.DeleteRunnersRunnerID(&runner.DeleteRunnersRunnerIDParams{
			RunnerID: id,
		}, nil)
		if err != nil {
			return fmt.Errorf("error deleting runner out-of-band: %s", err.Error())
		}
		return nil
	}
}

func testAccRunnerImportID(n string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return "", fmt.Errorf("not found: %s", n)
		}

		return fmt.Sprintf("runner/%s", rs.Primary.Attributes["id"]), nil
	}
}

func TestAcc_RunnerResource_basic(t *testing.T) {
	nameSuffix := acctest.RandString(8)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckRunnerDestroy,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccRunnerConfig(nameSuffix, 1, true, false, `["linux", "production"]`),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccRunnerExists("semaphoreui_runner.test"),
					resource.TestCheckResourceAttr("semaphoreui_runner.test", "name", fmt.Sprintf("Test %s", nameSuffix)),
					resource.TestCheckResourceAttr("semaphoreui_runner.test", "max_parallel_tasks", "1"),
					resource.TestCheckResourceAttr("semaphoreui_runner.test", "active", "true"),
					resource.TestCheckResourceAttr("semaphoreui_runner.test", "is_default", "false"),
					resource.TestCheckResourceAttr("semaphoreui_runner.test", "tags.#", "2"),
					resource.TestCheckTypeSetElemAttr("semaphoreui_runner.test", "tags.*", "linux"),
					resource.TestCheckTypeSetElemAttr("semaphoreui_runner.test", "tags.*", "production"),
					resource.TestCheckResourceAttrSet("semaphoreui_runner.test", "id"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "semaphoreui_runner.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccRunnerImportID("semaphoreui_runner.test"),
			},
			// Update and Read testing
			{
				Config: testAccRunnerConfig(nameSuffix, 4, false, true, `["windows"]`),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccRunnerExists("semaphoreui_runner.test"),
					resource.TestCheckResourceAttr("semaphoreui_runner.test", "name", fmt.Sprintf("Test %s", nameSuffix)),
					resource.TestCheckResourceAttr("semaphoreui_runner.test", "max_parallel_tasks", "4"),
					resource.TestCheckResourceAttr("semaphoreui_runner.test", "active", "false"),
					resource.TestCheckResourceAttr("semaphoreui_runner.test", "is_default", "true"),
					resource.TestCheckResourceAttr("semaphoreui_runner.test", "tags.#", "1"),
					resource.TestCheckTypeSetElemAttr("semaphoreui_runner.test", "tags.*", "windows"),
				),
			},
			// Deletion is verified by CheckDestroy after the test completes.
		},
	})
}

// TestAcc_RunnerResource_disappears verifies that a runner deleted out-of-band
// (e.g. from the web UI) is detected as drift and planned for recreation,
// rather than failing the plan with a 404 error.
func TestAcc_RunnerResource_disappears(t *testing.T) {
	nameSuffix := acctest.RandString(8)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckRunnerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRunnerConfig(nameSuffix, 1, true, false, `["linux"]`),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccRunnerExists("semaphoreui_runner.test"),
					testAccDeleteRunnerOutOfBand("semaphoreui_runner.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}
