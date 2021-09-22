package iterative

import (
	"context"
	"fmt"
	"log"
	"net"
	"terraform-provider-iterative/iterative/utils"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceDVCRun() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDVCRunCreate,
		DeleteContext: resourceRunnerDelete,
		ReadContext:   resourceMachineRead,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"repo": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"dvc_ver": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				Default:  "latest",
			},
			"labels": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  "dvc",
			},
			"idle_timeout": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
				Default:  300,
			},
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  "",
			},
			"single": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
				Default:  false,
			},
			"cloud": &schema.Schema{
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
				Default:  "",
			},
			"region": &schema.Schema{
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
				Default:  "us-west",
			},
			"image": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"spot": &schema.Schema{
				Type:     schema.TypeBool,
				ForceNew: true,
				Optional: true,
				Default:  false,
			},
			"spot_price": &schema.Schema{
				Type:     schema.TypeFloat,
				ForceNew: true,
				Optional: true,
				Default:  -1,
			},
			"instance_type": &schema.Schema{
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
				Default:  "m",
			},
			"instance_hdd_size": &schema.Schema{
				Type:     schema.TypeInt,
				ForceNew: true,
				Optional: true,
				Default:  35,
			},
			"instance_gpu": &schema.Schema{
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
				Default:  "",
			},
			"instance_ip": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"instance_launch_time": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"ssh_public": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"ssh_private": &schema.Schema{
				Type:      schema.TypeString,
				ForceNew:  true,
				Optional:  true,
				Default:   "",
				Sensitive: true,
			},
			"startup_script": &schema.Schema{
				Type:      schema.TypeString,
				ForceNew:  true,
				Optional:  true,
				Default:   "",
				Sensitive: true,
			},
			"aws_security_group": &schema.Schema{
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
				Default:  "",
			},
			"metadata": &schema.Schema{
				Type:     schema.TypeMap,
				ForceNew: true,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func dvcRunProvisioner(d *schema.ResourceData) (string, error) {
	return "", nil
}

func resourceDVCRunCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	utils.SetId(d)

	startupScript, err := dvcRunProvisioner(d)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Error generating startup script: %s", err),
		})
		return diags
	}

	d.Set("startup_script", startupScript)
	if d.Get("instance_gpu") == "tesla" {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  fmt.Sprintf("GPU model 'tesla' has been deprecated; please use 'v100' instead"),
		})
		d.Set("instance_gpu", "v100")
	}

	if len(d.Get("cloud").(string)) == 0 {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Local runner not yet implemented"),
		})
	} else {
		diags = resourceMachineCreate(ctx, d, m)
	}

	if diags.HasError() {
		return diags
	}
	log.Printf("[DEBUG] Instance address: %#v", d.Get("instance_ip"))

	var logError error
	var logEvents string
	err = resource.Retry(d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		switch cloud := d.Get("cloud").(string); cloud {
		case "kubernetes":
			logEvents, logError = resourceMachineLogs(ctx, d, m)
		default:
			logEvents, logError = utils.RunCommand("journalctl --unit dvc --no-pager",
				2*time.Second,
				net.JoinHostPort(d.Get("instance_ip").(string), "22"),
				"ubuntu",
				d.Get("ssh_private").(string))
		}

		log.Printf("[DEBUG] Collected log events: %#v", logEvents)
		log.Printf("[DEBUG] Connection errors: %#v", logError)

		if logError != nil {
			return resource.RetryableError(fmt.Errorf("Waiting for the machine to accept connections... %s", logError))
		} else if utils.HasStatus(logEvents, "terminated") {
			return resource.NonRetryableError(fmt.Errorf("Failed to launch the runner!"))
		} else if utils.HasStatus(logEvents, "ready") {
			return nil
		}

		return resource.RetryableError(fmt.Errorf("Waiting for the runner to be ready..."))
	})

	if logError != nil {
		logEvents += "\n" + logError.Error()
	}

	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Error checking the runner status"),
			Detail:   logEvents,
		})
	}

	return diags
}
