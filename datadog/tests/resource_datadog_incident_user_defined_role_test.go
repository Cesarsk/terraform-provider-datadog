package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccDatadogIncidentUserDefinedRole_Basic(t *testing.T) {
	skipIfNoCassette(t)
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	roleName := fmt.Sprintf("test_udr_basic_%d", clockFromContext(ctx).Now().Unix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogIncidentUserDefinedRoleDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogIncidentUserDefinedRoleConfig(roleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogIncidentUserDefinedRoleExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_incident_user_defined_role.foo", "name", roleName),
					resource.TestCheckResourceAttr(
						"datadog_incident_user_defined_role.foo", "description", "Created by terraform"),
					resource.TestCheckResourceAttr(
						"datadog_incident_user_defined_role.foo", "policy.is_single", "true"),
					resource.TestCheckResourceAttrSet(
						"datadog_incident_user_defined_role.foo", "incident_type"),
					resource.TestCheckResourceAttrSet(
						"datadog_incident_user_defined_role.foo", "created"),
				),
			},
			{
				Config: testAccCheckDatadogIncidentUserDefinedRoleConfigUpdated(roleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogIncidentUserDefinedRoleExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_incident_user_defined_role.foo", "name", roleName+"-updated"),
					resource.TestCheckResourceAttr(
						"datadog_incident_user_defined_role.foo", "description", "Updated by terraform"),
					resource.TestCheckResourceAttr(
						"datadog_incident_user_defined_role.foo", "policy.is_single", "false"),
				),
			},
		},
	})
}

func testAccCheckDatadogIncidentUserDefinedRoleConfig(roleName string) string {
	return fmt.Sprintf(`
resource "datadog_incident_type" "test" {
  name        = "%s-it"
  description = "Test incident type for user-defined role"
}

resource "datadog_incident_user_defined_role" "foo" {
  name          = "%s"
  description   = "Created by terraform"
  incident_type = datadog_incident_type.test.id
  policy = {
    is_single = true
  }
}`, roleName, roleName)
}

func testAccCheckDatadogIncidentUserDefinedRoleConfigUpdated(roleName string) string {
	return fmt.Sprintf(`
resource "datadog_incident_type" "test" {
  name        = "%s-it"
  description = "Test incident type for user-defined role"
}

resource "datadog_incident_user_defined_role" "foo" {
  name          = "%s-updated"
  description   = "Updated by terraform"
  incident_type = datadog_incident_type.test.id
  policy = {
    is_single = false
  }
}`, roleName, roleName)
}

func testAccCheckDatadogIncidentUserDefinedRoleExists(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		return incidentUserDefinedRoleExistsHelper(auth, s, apiInstances)
	}
}

func testAccCheckDatadogIncidentUserDefinedRoleDestroy(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		return incidentUserDefinedRoleDestroyHelper(auth, s, apiInstances)
	}
}

func incidentUserDefinedRoleExistsHelper(ctx context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	api := apiInstances.GetIncidentsApiV2()
	for _, r := range s.RootModule().Resources {
		if r.Type != "datadog_incident_user_defined_role" {
			continue
		}
		id, err := uuid.Parse(r.Primary.ID)
		if err != nil {
			return err
		}

		_, httpResp, err := api.GetIncidentUserDefinedRole(ctx, id)
		if err != nil {
			return utils.TranslateClientError(err, httpResp, "error retrieving incident user-defined role")
		}
	}
	return nil
}

func incidentUserDefinedRoleDestroyHelper(ctx context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	api := apiInstances.GetIncidentsApiV2()
	for _, r := range s.RootModule().Resources {
		if r.Type != "datadog_incident_user_defined_role" {
			continue
		}
		id, err := uuid.Parse(r.Primary.ID)
		if err != nil {
			return err
		}

		_, httpResp, err := api.GetIncidentUserDefinedRole(ctx, id)
		if err != nil {
			if httpResp != nil && httpResp.StatusCode == 404 {
				continue
			}
			return utils.TranslateClientError(err, httpResp, "error retrieving incident user-defined role")
		}
		return fmt.Errorf("incident user-defined role still exists")
	}
	return nil
}
