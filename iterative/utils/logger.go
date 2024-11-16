package utils

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/acarl005/stripansi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/sirupsen/logrus"
)

var colors = map[string]int{
	"DEBUG":      34,
	"INFO":       36,
	"WARNING":    33,
	"ERROR":      31,
	"FATAL":      31,
	"SUCCESS":    32,
	"foreground": 35,
}

type TpiFormatter struct{}

func (f *TpiFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	levelText := strings.ToUpper(entry.Level.String())
	levelColor := colors[levelText]
	message := entry.Message

	if d, ok := entry.Data["d"].(*schema.ResourceData); ok {
		switch message {
		case "instance":
			message = formatSchemaInstance(d)
		case "status":
			message = formatSchemaStatus(d)
		case "logs":
			message = formatSchemaLogs(d)
		default:
			return nil, errors.New("wrong schema logging mode")
		}
	}

	newPrefix := fmt.Sprintf("\x1b[%dmTPI [%s]\x1b[0m", levelColor, levelText)
	return []byte(hideUnwantedPrefix(levelText, newPrefix, message)), nil
}

func hideUnwantedPrefix(levelText, newPrefix, message string) string {
	timeString := time.Now().Format("2006-01-02T15:04:05.000Z0700")
	unwantedPrefixLength := len(fmt.Sprintf("%s [%s] provider.terraform-provider-iterative: [%[2]s]", timeString, levelText))

	var output string
	for _, line := range strings.Split(message, "\n") {
		formattedLine := fmt.Sprintf("[%s]\r%s %s", levelText, newPrefix, line)
		padding := strings.Repeat(" ", int(math.Max(float64(unwantedPrefixLength-len([]rune(stripansi.Strip(line)))), 0)))
		output += formattedLine + padding + "\n"
	}

	return output
}

func formatSchemaInstance(d *schema.ResourceData) string {
	cloud := d.Get("cloud").(string)
	machine := d.Get("machine").(string)
	region := d.Get("region").(string)
	spot := d.Get("spot").(float64)

	spottext := ""
	if spot > 0 {
		spottext = fmt.Sprintf("(Spot %f/h)", spot)
	}
	return fmt.Sprintf("%s %s%s in %s", cloud, machine, spottext, region)
}

func formatSchemaStatus(d *schema.ResourceData) string {
	status := d.Get("status").(map[string]interface{})

	message := fmt.Sprintf("\x1b[%dmStatus: queued \x1b[1m•\x1b[0m", colors["DEBUG"])
	if status["succeeded"] != nil && status["succeeded"].(int) >= d.Get("parallelism").(int) {
		message = fmt.Sprintf("\x1b[%dmStatus: completed successfully \x1b[1m•\x1b[0m", colors["SUCCESS"])
	}
	if status["failed"] != nil && status["failed"].(int) > 0 {
		message = fmt.Sprintf("\x1b[%dmStatus: completed with errors \x1b[1m•\x1b[0m", colors["ERROR"])
	}
	if status["timeout"] != nil && status["timeout"].(int) > 0 {
		message = fmt.Sprintf("\x1b[%dmStatus: stopped via timeout \x1b[1m•\x1b[0m", colors["WARNING"])
	}
	if status["running"] != nil && status["running"].(int) >= d.Get("parallelism").(int) {
		message = fmt.Sprintf("\x1b[%dmStatus: running \x1b[1m•\x1b[0m", colors["WARNING"])
	}
	return message
}

func formatSchemaLogs(d *schema.ResourceData) string {
	logs := d.Get("logs").([]interface{})

	message := ""
	for index, log := range logs {
		prefix := fmt.Sprintf("\n\x1b[%dmLOG %d >> ", colors["foreground"], index)
		message += strings.Trim(strings.ReplaceAll("\n"+strings.Trim(log.(string), "\n"), "\n", prefix), "\n")
		if index+1 < len(logs) {
			message += "\n"
		}
	}
	return message
}
