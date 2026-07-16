package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// NOTE: Requires the regenerated datadog-api-client-go with the expanded
// PostmortemTemplate types (datadog-api-spec PR #6178). Record cassettes once the
// client is bumped.

func TestAccDatadogIncidentPostmortemTemplate_Basic(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogIncidentPostmortemTemplateDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogIncidentPostmortemTemplateConfig(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogIncidentPostmortemTemplateExists(accProvider, "datadog_incident_postmortem_template.foo"),
					resource.TestCheckResourceAttr("datadog_incident_postmortem_template.foo", "name", uniq),
					resource.TestCheckResourceAttr("datadog_incident_postmortem_template.foo", "location", "datadog_notebooks"),
					resource.TestCheckResourceAttrSet("datadog_incident_postmortem_template.foo", "incident_type"),
				),
			},
			{
				Config: testAccCheckDatadogIncidentPostmortemTemplateConfigUpdated(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogIncidentPostmortemTemplateExists(accProvider, "datadog_incident_postmortem_template.foo"),
					resource.TestCheckResourceAttr("datadog_incident_postmortem_template.foo", "name", uniq+"-updated"),
					resource.TestCheckResourceAttr("datadog_incident_postmortem_template.foo", "is_default", "true"),
				),
			},
			{
				ResourceName:      "datadog_incident_postmortem_template.foo",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckDatadogIncidentPostmortemTemplateConfig(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_incident_type" "foo" {
  name = "%s-type"
}

resource "datadog_incident_postmortem_template" "foo" {
  name          = "%s"
  content       = "# Overview\n{{incident.card}}\n"
  incident_type = datadog_incident_type.foo.id
}`, uniq, uniq)
}

func testAccCheckDatadogIncidentPostmortemTemplateConfigUpdated(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_incident_type" "foo" {
  name = "%s-type"
}

resource "datadog_incident_postmortem_template" "foo" {
  name          = "%s-updated"
  content       = "# Overview\n{{incident.card}}\n\n# Timeline\n{{incident.timeline}}\n"
  is_default    = true
  incident_type = datadog_incident_type.foo.id
}`, uniq, uniq)
}

func testAccCheckDatadogIncidentPostmortemTemplateExists(accProvider func() (*schemaProvider, error), resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		provider, _ := accProvider()
		apiInstances := provider.DatadogApiInstances
		auth := provider.Auth

		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource %s not found", resourceName)
		}
		id, err := parseUUID(rs.Primary.ID)
		if err != nil {
			return err
		}
		_, httpResp, err := apiInstances.GetIncidentsApiV2().GetIncidentPostmortemTemplate(auth, id)
		if err != nil {
			return fmt.Errorf("received an error retrieving incident postmortem template: %s (HTTP %v)", err, httpResp)
		}
		return nil
	}
}

func testAccCheckDatadogIncidentPostmortemTemplateDestroy(accProvider func() (*schemaProvider, error)) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		provider, _ := accProvider()
		apiInstances := provider.DatadogApiInstances
		auth := provider.Auth

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "datadog_incident_postmortem_template" {
				continue
			}
			id, err := parseUUID(rs.Primary.ID)
			if err != nil {
				return err
			}
			_, httpResp, err := apiInstances.GetIncidentsApiV2().GetIncidentPostmortemTemplate(auth, id)
			if err != nil {
				if httpResp != nil && httpResp.StatusCode == 404 {
					continue
				}
				return fmt.Errorf("received an error retrieving incident postmortem template: %s", err)
			}
			return fmt.Errorf("incident postmortem template still exists")
		}
		return nil
	}
}
