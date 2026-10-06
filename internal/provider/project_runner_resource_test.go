package provider

import (
	"fmt"
	"strconv"
	"strings"
	"terraform-provider-semaphoreui/semaphoreui/client/project"
	"terraform-provider-semaphoreui/semaphoreui/client/runner"
	"terraform-provider-semaphoreui/semaphoreui/models"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// testAccPreCheckProjectRunner skips the test when the SemaphoreUI server under
// test does not allow project-scoped runners. They are a paid-plan feature: the
// community edition (used by CI) answers project runner creation with a 403
// "Your plan does not allow adding more runners." Global runners are unaffected.
func testAccPreCheckProjectRunner(t *testing.T) {
	testAccPreCheck(t)

	client := testClient()
	proj, err := client.Project.PostProjects(&project.PostProjectsParams{
		Project: &models.ProjectRequest{Name: "runner-precheck-" + acctest.RandString(8)},
	}, nil)
	if err != nil {
		// Could not probe; let the test run and surface any real error.
		return
	}
	projectID := proj.Payload.ID
	defer func() {
		_, _ = client.Project.DeleteProjectProjectID(&project.DeleteProjectProjectIDParams{
			ProjectID: projectID,
		}, nil)
	}()

	created, err := client.Runner.PostProjectProjectIDRunners(&runner.PostProjectProjectIDRunnersParams{
		ProjectID: projectID,
		Runner:    &models.RunnerRequest{ProjectID: projectID, Name: "precheck"},
	}, nil)
	if err != nil {
		if strings.Contains(err.Error(), "403") {
			t.Skip("skipping: SemaphoreUI plan does not allow project runners")
		}
		return
	}
	_, _ = client.Runner.DeleteProjectProjectIDRunnersRunnerID(&runner.DeleteProjectProjectIDRunnersRunnerIDParams{
		ProjectID: projectID,
		RunnerID:  created.Payload.ID,
	}, nil)
}

func testAccProjectRunnerExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		if rs.Primary.Attributes["id"] == "" {
			return fmt.Errorf("no ID is set")
		}
		if rs.Primary.Attributes["project_id"] == "" {
			return fmt.Errorf("no ProjectID is set")
		}

		id, _ := strconv.ParseInt(rs.Primary.Attributes["id"], 10, 64)
		projectId, _ := strconv.ParseInt(rs.Primary.Attributes["project_id"], 10, 64)

		_, err := testClient().Runner.GetProjectProjectIDRunnersRunnerID(&runner.GetProjectProjectIDRunnersRunnerIDParams{
			ProjectID: projectId,
			RunnerID:  id,
		}, nil)
		if err != nil {
			return fmt.Errorf("error reading project runner: %s", err.Error())
		}

		return nil
	}
}

func testAccProjectRunnerProjectConfig(nameSuffix string) string {
	return fmt.Sprintf(`
resource "semaphoreui_project" "test" {
  name = "test-%[1]s"
}
`, nameSuffix)
}

func testAccProjectRunnerConfig(nameSuffix string, maxParallelTasks int, active bool, isDefault bool, tags string) string {
	return fmt.Sprintf(`
%[1]s
resource "semaphoreui_project_runner" "test" {
  project_id         = semaphoreui_project.test.id
  name               = "Test %[2]s"
  max_parallel_tasks = %[3]d
  active             = %[4]t
  is_default         = %[5]t
  tags               = %[6]s
}`, testAccProjectRunnerProjectConfig(nameSuffix), nameSuffix, maxParallelTasks, active, isDefault, tags)
}

func testAccProjectRunnerImportID(n string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return "", fmt.Errorf("not found: %s", n)
		}

		return fmt.Sprintf("project/%[1]s/runner/%[2]s", rs.Primary.Attributes["project_id"], rs.Primary.Attributes["id"]), nil
	}
}

func TestAcc_ProjectRunnerResource_basic(t *testing.T) {
	nameSuffix := acctest.RandString(8)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheckProjectRunner(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccProjectRunnerConfig(nameSuffix, 1, true, false, `["linux", "production"]`),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccProjectRunnerExists("semaphoreui_project_runner.test"),
					resource.TestCheckResourceAttr("semaphoreui_project_runner.test", "name", fmt.Sprintf("Test %s", nameSuffix)),
					resource.TestCheckResourceAttr("semaphoreui_project_runner.test", "max_parallel_tasks", "1"),
					resource.TestCheckResourceAttr("semaphoreui_project_runner.test", "active", "true"),
					resource.TestCheckResourceAttr("semaphoreui_project_runner.test", "is_default", "false"),
					resource.TestCheckResourceAttr("semaphoreui_project_runner.test", "tags.#", "2"),
					resource.TestCheckTypeSetElemAttr("semaphoreui_project_runner.test", "tags.*", "linux"),
					resource.TestCheckTypeSetElemAttr("semaphoreui_project_runner.test", "tags.*", "production"),
					resource.TestCheckResourceAttrSet("semaphoreui_project_runner.test", "id"),
					resource.TestCheckResourceAttrSet("semaphoreui_project_runner.test", "project_id"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "semaphoreui_project_runner.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccProjectRunnerImportID("semaphoreui_project_runner.test"),
			},
			// Update and Read testing
			{
				Config: testAccProjectRunnerConfig(nameSuffix, 4, false, true, `["windows"]`),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccProjectRunnerExists("semaphoreui_project_runner.test"),
					resource.TestCheckResourceAttr("semaphoreui_project_runner.test", "name", fmt.Sprintf("Test %s", nameSuffix)),
					resource.TestCheckResourceAttr("semaphoreui_project_runner.test", "max_parallel_tasks", "4"),
					resource.TestCheckResourceAttr("semaphoreui_project_runner.test", "active", "false"),
					resource.TestCheckResourceAttr("semaphoreui_project_runner.test", "is_default", "true"),
					resource.TestCheckResourceAttr("semaphoreui_project_runner.test", "tags.#", "1"),
					resource.TestCheckTypeSetElemAttr("semaphoreui_project_runner.test", "tags.*", "windows"),
				),
			},
			// Delete testing
			{
				Config: testAccProjectRunnerProjectConfig(nameSuffix),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccResourceNotExists("semaphoreui_project_runner.test"),
				),
			},
		},
	})
}
