package provider

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func testAccProjectEnvironmentDataSourceConfig() string {
	return `
resource "semaphoreui_project" "test" {
  name = "Project 1"
}

resource "semaphoreui_project_environment" "test" {
  project_id = semaphoreui_project.test.id
  name       = "Test Environment"

  # extraVars
  variables = {
    key1 = "value1"
    key2 = "value2"
  }

  # environment variables
  environment = {
    KEY1 = "value1"
    KEY2 = "value2"
  }

  # secrets
  secrets = [{
    # extraVar Secret
    name  = "key3"
    type  = "var"
    value = "value3"
    }, {
    # environment Secret
    name  = "KEY4"
    type  = "env"
    value = "value4"
  }]
}

data "semaphoreui_project_environment" "test" {
  project_id = semaphoreui_project.test.id
  id         = semaphoreui_project_environment.test.id
  depends_on = [semaphoreui_project_environment.test]
}

data "semaphoreui_project_environment" "by_name" {
  project_id = semaphoreui_project.test.id
  name       = semaphoreui_project_environment.test.name
  depends_on = [semaphoreui_project_environment.test]
}`
}

func testAccProjectEnvironmentDataSourceConfigNameNotFound() string {
	return `
resource "semaphoreui_project" "test" {
  name = "Project 1"
}

data "semaphoreui_project_environment" "missing" {
  project_id = semaphoreui_project.test.id
  name       = "Does Not Exist"
}`
}

func testAccProjectEnvironmentDataSourceConfigIDAndName() string {
	return `
data "semaphoreui_project_environment" "test" {
  project_id = 1
  id         = 1
  name       = "Test Environment"
}`
}

func testAccProjectEnvironmentDataSourceConfigNoIDOrName() string {
	return `
data "semaphoreui_project_environment" "test" {
  project_id = 1
}`
}

func TestAcc_ProjectEnvironmentDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testAccProjectEnvironmentDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.semaphoreui_project_environment.test", "name", "Test Environment"),
					resource.TestCheckResourceAttr("data.semaphoreui_project_environment.test", "variables.%", "2"),
					resource.TestCheckResourceAttr("data.semaphoreui_project_environment.test", "variables.key1", "value1"),
					resource.TestCheckResourceAttr("data.semaphoreui_project_environment.test", "variables.key2", "value2"),
					resource.TestCheckResourceAttr("data.semaphoreui_project_environment.test", "environment.%", "2"),
					resource.TestCheckResourceAttr("data.semaphoreui_project_environment.test", "environment.KEY1", "value1"),
					resource.TestCheckResourceAttr("data.semaphoreui_project_environment.test", "environment.KEY2", "value2"),
					resource.TestCheckResourceAttr("data.semaphoreui_project_environment.test", "secrets.#", "2"),
					resource.TestCheckResourceAttr("data.semaphoreui_project_environment.test", "secrets.0.name", "key3"),
					resource.TestCheckResourceAttr("data.semaphoreui_project_environment.test", "secrets.0.value", ""),
					resource.TestCheckResourceAttr("data.semaphoreui_project_environment.test", "secrets.0.type", "var"),
					resource.TestCheckResourceAttr("data.semaphoreui_project_environment.test", "secrets.1.name", "KEY4"),
					resource.TestCheckResourceAttr("data.semaphoreui_project_environment.test", "secrets.1.value", ""),
					resource.TestCheckResourceAttr("data.semaphoreui_project_environment.test", "secrets.1.type", "env"),
					// Lookup by name resolves to the same environment
					resource.TestCheckResourceAttrPair("data.semaphoreui_project_environment.by_name", "id", "semaphoreui_project_environment.test", "id"),
					resource.TestCheckResourceAttr("data.semaphoreui_project_environment.by_name", "name", "Test Environment"),
					resource.TestCheckResourceAttr("data.semaphoreui_project_environment.by_name", "variables.%", "2"),
					resource.TestCheckResourceAttr("data.semaphoreui_project_environment.by_name", "environment.%", "2"),
					resource.TestCheckResourceAttr("data.semaphoreui_project_environment.by_name", "secrets.#", "2"),
				),
			},
		},
	})
}

func TestAcc_ProjectEnvironmentDataSource_nameNotFound(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccProjectEnvironmentDataSourceConfigNameNotFound(),
				ExpectError: regexp.MustCompile(`project environment with name "Does Not Exist" not found`),
			},
		},
	})
}

func TestAcc_ProjectEnvironmentDataSource_exactlyOneOfIDOrName(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccProjectEnvironmentDataSourceConfigIDAndName(),
				ExpectError: regexp.MustCompile(`Invalid Attribute Combination`),
			},
			{
				Config:      testAccProjectEnvironmentDataSourceConfigNoIDOrName(),
				ExpectError: regexp.MustCompile(`Invalid Attribute Combination`),
			},
		},
	})
}

func testAccProjectEnvironmentDataSourceConfigDuplicateName() string {
	return `
resource "semaphoreui_project" "test" {
  name = "Project 1"
}

resource "semaphoreui_project_environment" "first" {
  project_id = semaphoreui_project.test.id
  name       = "Duplicate"
}

resource "semaphoreui_project_environment" "second" {
  project_id = semaphoreui_project.test.id
  name       = "Duplicate"
}

data "semaphoreui_project_environment" "test" {
  project_id = semaphoreui_project.test.id
  name       = "Duplicate"
  depends_on = [
    semaphoreui_project_environment.first,
    semaphoreui_project_environment.second,
  ]
}`
}

func TestAcc_ProjectEnvironmentDataSource_duplicateName(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccProjectEnvironmentDataSourceConfigDuplicateName(),
				ExpectError: regexp.MustCompile(`multiple project environments with name "Duplicate" found`),
			},
		},
	})
}
