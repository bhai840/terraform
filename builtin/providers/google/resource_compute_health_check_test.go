package google

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"google.golang.org/api/compute/v1"
)

func TestAccComputeHealthCheck_basic(t *testing.T) {
	var healthCheck compute.HealthCheck

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckComputeHealthCheckDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccComputeHealthCheck_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeHealthCheckExists(
						"google_compute_health_check.foobar", &healthCheck),
					testAccCheckComputeHealthCheckThresholds(
						3, 3, &healthCheck),
				),
			},
		},
	})
}

func TestAccComputeHealthCheck_update(t *testing.T) {
	var healthCheck compute.HealthCheck

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckComputeHealthCheckDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccComputeHealthCheck_update1,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeHealthCheckExists(
						"google_compute_health_check.foobar", &healthCheck),
					testAccCheckComputeHealthCheckThresholds(
						2, 2, &healthCheck),
				),
			},
			resource.TestStep{
				Config: testAccComputeHealthCheck_update2,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeHealthCheckExists(
						"google_compute_health_check.foobar", &healthCheck),
					testAccCheckComputeHealthCheckThresholds(
						10, 10, &healthCheck),
				),
			},
		},
	})
}

func testAccCheckComputeHealthCheckDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "google_compute_health_check" {
			continue
		}

		_, err := config.clientCompute.HealthChecks.Get(
			config.Project, rs.Primary.ID).Do()
		if err == nil {
			return fmt.Errorf("HealthCheck still exists")
		}
	}

	return nil
}

func testAccCheckComputeHealthCheckExists(n string, healthCheck *compute.HealthCheck) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := testAccProvider.Meta().(*Config)

		found, err := config.clientCompute.HealthChecks.Get(
			config.Project, rs.Primary.ID).Do()
		if err != nil {
			return err
		}

		if found.Name != rs.Primary.ID {
			return fmt.Errorf("HealthCheck not found")
		}

		*healthCheck = *found

		return nil
	}
}

func testAccCheckComputeHealthCheckThresholds(healthy, unhealthy int64, healthCheck *compute.HealthCheck) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if healthCheck.HealthyThreshold != healthy {
			return fmt.Errorf("HealthyThreshold doesn't match: expected %d, got %d", healthy, healthCheck.HealthyThreshold)
		}

		if healthCheck.UnhealthyThreshold != unhealthy {
			return fmt.Errorf("UnhealthyThreshold doesn't match: expected %d, got %d", unhealthy, healthCheck.UnhealthyThreshold)
		}

		return nil
	}
}

var testAccComputeHealthCheck_basic = fmt.Sprintf(`
resource "google_compute_health_check" "foobar" {
	check_interval_sec = 3
	description = "Resource created for Terraform acceptance testing"
	healthy_threshold = 3
	name = "health-test-%s"
	timeout_sec = 2
	unhealthy_threshold = 3
	tcp_health_check {
		port = "80"
	}
}
`, acctest.RandString(10))

var testAccComputeHealthCheck_update1 = fmt.Sprintf(`
resource "google_compute_health_check" "foobar" {
	name = "Health-test-%s"
	description = "Resource created for Terraform acceptance testing"
	request_path = "/not_default"
}
`, acctest.RandString(10))

/* Change description, restore request_path to default, and change
* thresholds from defaults */
var testAccComputeHealthCheck_update2 = fmt.Sprintf(`
resource "google_compute_health_check" "foobar" {
	name = "Health-test-%s"
	description = "Resource updated for Terraform acceptance testing"
	healthy_threshold = 10
	unhealthy_threshold = 10
}
`, acctest.RandString(10))
