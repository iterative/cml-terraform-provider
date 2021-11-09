package kubernetes

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"regexp"
	"strconv"
	"time"

	"terraform-provider-iterative/task/kubernetes/client"
	"terraform-provider-iterative/task/kubernetes/resources"
	"terraform-provider-iterative/task/universal"
	"terraform-provider-iterative/task/universal/ssh"
)

func NewTask(ctx context.Context, cloud universal.Cloud, identifier string, task universal.Task) (*Task, error) {
	client, err := client.New(ctx, cloud, task.Tags)
	if err != nil {
		return nil, err
	}

	match := regexp.MustCompile(`^([^:]+):(\d+)(?::(.+))?$`).FindStringSubmatch(task.Environment.Directory)
	if match == nil {
		return nil, errors.New("directory specification for k8s is a bit different; see the documentation for more information")
	}

	storageClass := match[1]
	directory := match[3]
	size, err := strconv.Atoi(match[2])
	if err != nil {
		return nil, err
	}

	t := new(Task)
	t.Client = client
	t.Identifier = identifier
	t.Attributes.Task = task
	t.Attributes.Directory = directory
	t.Resources.PersistentVolumeClaim = resources.NewPersistentVolumeClaim(
		t.Client,
		t.Identifier,
		storageClass,
		uint64(size),
		t.Attributes.Task.Parallelism > 1,
	)
	t.Resources.Job = resources.NewJob(
		t.Client,
		t.Identifier,
		t.Resources.PersistentVolumeClaim,
		t.Attributes.Task,
	)
	return t, nil
}

type Task struct {
	Client     *client.Client
	Identifier string
	Attributes struct {
		universal.Task
		Directory string
	}
	DataSources struct{}
	Resources   struct {
		*resources.PersistentVolumeClaim
		*resources.Job
	}
}

func (t *Task) Create(ctx context.Context) error {
	log.Println("[INFO] Creating PersistentVolumeClaim...")
	if err := t.Resources.PersistentVolumeClaim.Create(ctx); err != nil {
		return err
	}
	log.Println("[INFO] Creating Job...")
	if err := t.Resources.Job.Create(ctx); err != nil {
		return err
	}
	log.Println("[INFO] Uploading Directory...")
	if t.Attributes.Directory != "" {
		if err := t.Push(ctx, t.Attributes.Directory, false); err != nil {
			return err
		}
	}
	log.Println("[INFO] Done!")
	t.Attributes.Task.Addresses = t.Resources.Job.Attributes.Addresses
	t.Attributes.Task.Status = t.Resources.Job.Attributes.Status
	t.Attributes.Task.Events = t.Resources.Job.Attributes.Events
	return nil
}

func (t *Task) Read(ctx context.Context) error {
	log.Println("[INFO] Reading PersistentVolumeClaim...")
	if err := t.Resources.PersistentVolumeClaim.Read(ctx); err != nil {
		return err
	}
	log.Println("[INFO] Reading Job...")
	if err := t.Resources.Job.Read(ctx); err != nil {
		return err
	}
	log.Println("[INFO] Done!")
	t.Attributes.Task.Addresses = t.Resources.Job.Attributes.Addresses
	t.Attributes.Task.Status = t.Resources.Job.Attributes.Status
	t.Attributes.Task.Events = t.Resources.Job.Attributes.Events
	return nil
}

func (t *Task) Delete(ctx context.Context) error {
	log.Println("[INFO] Downloading Directory...")
	if t.Attributes.Directory != "" && t.Read(ctx) == nil {
		log.Println("[INFO] Deleting completed Job...")
		if err := t.Resources.Job.Delete(ctx); err != nil {
			return err
		}
		log.Println("[INFO] Creating ephemeral Job to retrieve directory...")
		if err := t.Resources.Job.Create(ctx); err != nil {
			return err
		}
		if err := t.Pull(ctx, t.Attributes.Directory); err != nil {
			return err
		}
	}
	log.Println("[INFO] Deleting Job...")
	if err := t.Resources.Job.Delete(ctx); err != nil {
		return err
	}
	log.Println("[INFO] Deleting PersistentVolumeClaim...")
	if err := t.Resources.PersistentVolumeClaim.Delete(ctx); err != nil {
		return err
	}
	log.Println("[INFO] Done!")
	return nil
}

func (t *Task) Push(ctx context.Context, source string, unsafe bool) error {
	destination := "/task"
	waitSelector := fmt.Sprintf("controller-uid=%s", t.Resources.Job.Resource.GetObjectMeta().GetLabels()["controller-uid"])
	pod, err := resources.WaitForPods(ctx, t.Client, 1*time.Second, t.Client.Cloud.Timeouts.Create, t.Client.Namespace, waitSelector)
	if err != nil {
		return err
	}
	return resources.CopyFile(t.Client, source+"/.", fmt.Sprintf("%s/%s:%s", t.Client.Namespace, pod, destination), t.Resources.Job.Identifier)
}

func (t *Task) Pull(ctx context.Context, destination string) error {
	source := "/task/."
	waitSelector := fmt.Sprintf("controller-uid=%s", t.Resources.Job.Resource.GetObjectMeta().GetLabels()["controller-uid"])
	pod, err := resources.WaitForPods(ctx, t.Client, 1*time.Second, t.Client.Cloud.Timeouts.Delete, t.Client.Namespace, waitSelector)
	if err != nil {
		return err
	}
	return resources.CopyFile(t.Client, fmt.Sprintf("%s/%s:%s", t.Client.Namespace, pod, source), destination, t.Resources.Job.Identifier)
}

func (t *Task) Logs(ctx context.Context) ([]string, error) {
	return t.Resources.Job.Logs(ctx)
}

func (t *Task) Stop(ctx context.Context) error {
	return errors.New("unsupported operation: Stop is intended for VM orchestrators")
}

func (t *Task) GetAddresses(ctx context.Context) []net.IP {
	return t.Attributes.Addresses
}

func (t *Task) GetEvents(ctx context.Context) []universal.Event {
	return t.Attributes.Events
}

func (t *Task) GetStatus(ctx context.Context) map[string]int {
	return t.Attributes.Status
}

func (t *Task) GetKeyPair(ctx context.Context) (*ssh.DeterministicSSHKeyPair, error) {
	return nil, universal.NotFoundError
}

func (t *Task) GetIdentifier(ctx context.Context) string {
	return t.Identifier
}
