package utils

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/sirupsen/logrus"
)

func TestState(t *testing.T) {
	d := generateSchemaData(t, map[string]interface{}{
		"name":        "mytask",
		"parallelism": "1",
		"status": map[string]interface{}{
			"running": 0,
			"failed":  0,
		},
	})

	logger := logrus.WithFields(logrus.Fields{"d": d})
	logrus.SetFormatter(&TpiFormatter{})
	logger.Info("status")
}

func TestState2(t *testing.T) {
	d := generateSchemaData(t, map[string]interface{}{
		"name":        "mytask",
		"parallelism": "1",
		"status": map[string]interface{}{
			"running": 0,
			"failed":  1,
		},
	})

	logger := logrus.WithFields(logrus.Fields{"d": d})
	logrus.SetFormatter(&TpiFormatter{})
	logger.Info("status")
}

func TestState3(t *testing.T) {
	d := generateSchemaData(t, map[string]interface{}{
		"name":        "mytask",
		"parallelism": "1",
		"status": map[string]interface{}{
			"running":   0,
			"succeeded": 1,
		},
	})

	logger := logrus.WithFields(logrus.Fields{"d": d})
	logrus.SetFormatter(&TpiFormatter{})
	logger.Info("status")
}

func TestState4(t *testing.T) {
	d := generateSchemaData(t, map[string]interface{}{
		"name":        "mytask",
		"parallelism": "1",
		"status": map[string]interface{}{
			"running": 0,
			"timeout": 1,
		},
	})

	logger := logrus.WithFields(logrus.Fields{"d": d})
	logrus.SetFormatter(&TpiFormatter{})
	logger.Info("status")
}

func TestLogs(t *testing.T) {
	logs := make([]interface{}, 0)
	logs = append(logs, "-- Logs begin at Tue 2022-03-01 12:25:09 UTC, end at Tue 2022-03-01 12:30:30 UTC. --\nMar 01 12:25:50 tpi000000 systemd[1]: Started tpi-task.service.\nMar 01 12:25:50 tpi000000 sudo[1706]:     root : TTY=unknown ; PWD=/tmp/tpi-task ; USER=root ; COMMAND=/usr/bin/apt update\nMar 01 12:25:50 tpi000000 sudo[1706]: pam_unix(sudo:session): session opened for user root by (uid=0)\nMar 01 12:25:50 tpi000000 tpi-task[1711]: WARNING: apt does not have a stable CLI interface. Use with caution in scripts.\nMar 01 12:25:50 tpi000000 tpi-task[1711]: Hit:1 http://azure.archive.ubuntu.com/ubuntu focal InRelease\nMar 01 12:25:50 tpi000000 tpi-task[1711]: Get:2 http://azure.archive.ubuntu.com/ubuntu focal-updates InRelease [114 kB]\nMar 01 12:25:50 tpi000000 tpi-task[1711]: Get:3 http://azure.archive.ubuntu.com/ubuntu focal-backports InRelease [108 kB]\nMar 01 12:25:51 tpi000000 tpi-task[1711]: Get:4 http://security.ubuntu.com/ubuntu focal-security InRelease [114 kB]\nMar 01 12:25:51 tpi000000 tpi-task[1711]: Get:5 http://azure.archive.ubuntu.com/ubuntu focal/universe amd64 Packages [8628 kB]\n")

	d := generateSchemaData(t, map[string]interface{}{
		"name": "mytask",
		"logs": logs,
	})

	logger := logrus.WithFields(logrus.Fields{"d": d})
	logrus.SetFormatter(&TpiFormatter{})
	logger.Info("logs")
}

func TestMachine(t *testing.T) {
	d := generateSchemaData(t, map[string]interface{}{
		"name":    "mytask",
		"cloud":   "aws",
		"machine": "t2.micro",
		"spot":    0.2,
		"region":  "us-west",
	})

	logger := logrus.WithFields(logrus.Fields{"d": d})
	logrus.SetFormatter(&TpiFormatter{})
	logger.Info("instance")
}

func generateSchemaData(t *testing.T, raw map[string]interface{}) *schema.ResourceData {
	sch := map[string]*schema.Schema{
		"name": {
			Type:     schema.TypeString,
			ForceNew: true,
			Optional: true,
		},
		"cloud": {
			Type:     schema.TypeString,
			ForceNew: true,
			Required: true,
		},
		"region": {
			Type:     schema.TypeString,
			ForceNew: true,
			Optional: true,
			Default:  "us-west",
		},
		"machine": {
			Type:     schema.TypeString,
			ForceNew: true,
			Optional: true,
			Default:  "m",
		},
		"disk_size": {
			Type:     schema.TypeInt,
			ForceNew: true,
			Optional: true,
			Default:  30,
		},
		"spot": {
			Type:     schema.TypeFloat,
			ForceNew: true,
			Optional: true,
			Default:  -1,
		},
		"image": {
			Type:     schema.TypeString,
			ForceNew: true,
			Optional: true,
			Default:  "ubuntu",
		},
		"addresses": {
			Type:     schema.TypeList,
			Computed: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"status": {
			Type:     schema.TypeMap,
			Computed: true,
			Elem: &schema.Schema{
				Type: schema.TypeInt,
			},
		},
		"events": {
			Type:     schema.TypeList,
			Computed: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"logs": {
			Type:     schema.TypeList,
			Computed: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"parallelism": {
			Type:     schema.TypeInt,
			ForceNew: true,
			Optional: true,
			Default:  1,
		},
		"script": {
			Type:     schema.TypeString,
			ForceNew: true,
			Required: true,
		},
		"workdir": {
			Optional: true,
			Type:     schema.TypeSet,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"input": {
						Type:     schema.TypeString,
						ForceNew: true,
						Optional: true,
						Default:  "",
					},
					"output": {
						Type:     schema.TypeString,
						ForceNew: false,
						Optional: true,
						Default:  "",
					},
				},
			},
		},
	}

	return schema.TestResourceDataRaw(t, sch, raw)
}
