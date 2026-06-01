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

func testAccRunnerConfig(nameSuffix string, maxParallelTasks int, active bool, tags string) string {
	return fmt.Sprintf(`
resource "semaphoreui_runner" "test" {
  name               = "Test %[1]s"
  max_parallel_tasks = %[2]d
  active             = %[3]t
  tags               = %[4]s
}`, nameSuffix, maxParallelTasks, active, tags)
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
				Config: testAccRunnerConfig(nameSuffix, 1, true, `["linux", "production"]`),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccRunnerExists("semaphoreui_runner.test"),
					resource.TestCheckResourceAttr("semaphoreui_runner.test", "name", fmt.Sprintf("Test %s", nameSuffix)),
					resource.TestCheckResourceAttr("semaphoreui_runner.test", "max_parallel_tasks", "1"),
					resource.TestCheckResourceAttr("semaphoreui_runner.test", "active", "true"),
					resource.TestCheckResourceAttr("semaphoreui_runner.test", "tags.#", "2"),
					resource.TestCheckResourceAttr("semaphoreui_runner.test", "tags.0", "linux"),
					resource.TestCheckResourceAttr("semaphoreui_runner.test", "tags.1", "production"),
					resource.TestCheckResourceAttrSet("semaphoreui_runner.test", "id"),
					resource.TestCheckResourceAttrSet("semaphoreui_runner.test", "registration_token"),
				),
			},
			// ImportState testing. The registration token is only returned at creation, so
			// it is not present after import and must be ignored on verification.
			{
				ResourceName:            "semaphoreui_runner.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateIdFunc:       testAccRunnerImportID("semaphoreui_runner.test"),
				ImportStateVerifyIgnore: []string{"registration_token"},
			},
			// Update and Read testing
			{
				Config: testAccRunnerConfig(nameSuffix, 4, false, `["windows"]`),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccRunnerExists("semaphoreui_runner.test"),
					resource.TestCheckResourceAttr("semaphoreui_runner.test", "name", fmt.Sprintf("Test %s", nameSuffix)),
					resource.TestCheckResourceAttr("semaphoreui_runner.test", "max_parallel_tasks", "4"),
					resource.TestCheckResourceAttr("semaphoreui_runner.test", "active", "false"),
					resource.TestCheckResourceAttr("semaphoreui_runner.test", "tags.#", "1"),
					resource.TestCheckResourceAttr("semaphoreui_runner.test", "tags.0", "windows"),
				),
			},
			// Deletion is verified by CheckDestroy after the test completes.
		},
	})
}
