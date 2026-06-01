package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func testAccRunnerDataSourceConfigByID() string {
	return `
resource "semaphoreui_runner" "test" {
  name               = "Test Global Runner"
  max_parallel_tasks = 2
  tags               = ["linux"]
}

data "semaphoreui_runner" "test" {
  id         = semaphoreui_runner.test.id
  depends_on = [semaphoreui_runner.test]
}`
}

func testAccRunnerDataSourceConfigByName() string {
	return `
resource "semaphoreui_runner" "test" {
  name               = "Test Global Runner"
  max_parallel_tasks = 2
  tags               = ["linux"]
}

data "semaphoreui_runner" "test" {
  name       = "Test Global Runner"
  depends_on = [semaphoreui_runner.test]
}`
}

func TestAcc_RunnerDataSource_basicID(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRunnerDataSourceConfigByID(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.semaphoreui_runner.test", "name", "Test Global Runner"),
					resource.TestCheckResourceAttr("data.semaphoreui_runner.test", "max_parallel_tasks", "2"),
					resource.TestCheckResourceAttr("data.semaphoreui_runner.test", "tags.#", "1"),
					resource.TestCheckTypeSetElemAttr("data.semaphoreui_runner.test", "tags.*", "linux"),
					resource.TestCheckResourceAttrSet("data.semaphoreui_runner.test", "id"),
				),
			},
		},
	})
}

func TestAcc_RunnerDataSource_basicName(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRunnerDataSourceConfigByName(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.semaphoreui_runner.test", "name", "Test Global Runner"),
					resource.TestCheckResourceAttr("data.semaphoreui_runner.test", "max_parallel_tasks", "2"),
					resource.TestCheckResourceAttr("data.semaphoreui_runner.test", "tags.#", "1"),
					resource.TestCheckTypeSetElemAttr("data.semaphoreui_runner.test", "tags.*", "linux"),
					resource.TestCheckResourceAttrSet("data.semaphoreui_runner.test", "id"),
				),
			},
		},
	})
}
