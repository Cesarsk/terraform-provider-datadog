package test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// NOTE: Requires the regenerated datadog-api-client-go with the expanded
// PostmortemTemplate types (datadog-api-spec PR #6178). Record cassettes once the
// client is bumped.

func TestAccDatadogIncidentPostmortemTemplate_Basic(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	name := fmt.Sprintf("test-postmortem-template-%d", clockFromContext(ctx).Now().Unix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogIncidentPostmortemTemplateDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogIncidentPostmortemTemplateConfig(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogIncidentPostmortemTemplateExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr("datadog_incident_postmortem_template.foo", "name", name),
					resource.TestCheckResourceAttr("datadog_incident_postmortem_template.foo", "location", "datadog_notebooks"),
					resource.TestCheckResourceAttrSet("datadog_incident_postmortem_template.foo", "id"),
					resource.TestCheckResourceAttrSet("datadog_incident_postmortem_template.foo", "incident_type"),
					resource.TestCheckResourceAttrSet("datadog_incident_postmortem_template.foo", "created"),
					resource.TestCheckResourceAttrSet("datadog_incident_postmortem_template.foo", "modified"),
				),
			},
			{
				Config: testAccCheckDatadogIncidentPostmortemTemplateConfigUpdated(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogIncidentPostmortemTemplateExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr("datadog_incident_postmortem_template.foo", "name", name+"-updated"),
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

// TestAccDatadogIncidentPostmortemTemplate_LocationSettingsMismatch asserts the
// plan-time ValidateConfig guard rejects settings that don't match the location
// before any API call is made.
func TestAccDatadogIncidentPostmortemTemplate_LocationSettingsMismatch(t *testing.T) {
	t.Parallel()
	ctx, _, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	name := fmt.Sprintf("test-postmortem-template-%d", clockFromContext(ctx).Now().Unix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccCheckDatadogIncidentPostmortemTemplateConfigMismatch(name),
				ExpectError: regexp.MustCompile(`google_docs_postmortem_settings may only be set when location is "google_docs"`),
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

func testAccCheckDatadogIncidentPostmortemTemplateConfigMismatch(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_incident_type" "foo" {
  name = "%s-type"
}

resource "datadog_incident_postmortem_template" "foo" {
  name          = "%s"
  location      = "datadog_notebooks"
  incident_type = datadog_incident_type.foo.id

  google_docs_postmortem_settings {
    account_id       = "123456"
    parent_folder_id = "789012"
  }
}`, uniq, uniq)
}

func testAccCheckDatadogIncidentPostmortemTemplateExists(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth
		return incidentPostmortemTemplateExistsHelper(auth, s, apiInstances)
	}
}

func testAccCheckDatadogIncidentPostmortemTemplateDestroy(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth
		return incidentPostmortemTemplateDestroyHelper(auth, s, apiInstances)
	}
}

func incidentPostmortemTemplateExistsHelper(ctx context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	api := apiInstances.GetIncidentsApiV2()
	for _, r := range s.RootModule().Resources {
		if r.Type != "datadog_incident_postmortem_template" {
			continue
		}
		_, httpResp, err := api.GetIncidentPostmortemTemplate(ctx, parseUUID(r.Primary.ID))
		if err != nil {
			return utils.TranslateClientError(err, httpResp, "error retrieving incident postmortem template")
		}
	}
	return nil
}

func incidentPostmortemTemplateDestroyHelper(ctx context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	api := apiInstances.GetIncidentsApiV2()
	for _, r := range s.RootModule().Resources {
		if r.Type != "datadog_incident_postmortem_template" {
			continue
		}
		_, httpResp, err := api.GetIncidentPostmortemTemplate(ctx, parseUUID(r.Primary.ID))
		if err != nil {
			if httpResp != nil && httpResp.StatusCode == 404 {
				continue
			}
			return utils.TranslateClientError(err, httpResp, "error retrieving incident postmortem template")
		}
		return fmt.Errorf("incident postmortem template still exists")
	}
	return nil
}
