package iterative

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		ResourcesMap: map[string]*schema.Resource{
			"iterative_machine":    resourceMachine(),
			"iterative_cml_runner": resourceRunner(),
			"iterative_dvc_run":    resourceDVCRun(),
		},
		DataSourcesMap: map[string]*schema.Resource{},
	}
}
