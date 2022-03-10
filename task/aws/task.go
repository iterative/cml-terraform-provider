package aws

import (
	"context"
	"net"

	"github.com/sirupsen/logrus"

	"terraform-provider-iterative/task/aws/client"
	"terraform-provider-iterative/task/aws/resources"
	"terraform-provider-iterative/task/common"
	"terraform-provider-iterative/task/common/machine"
	"terraform-provider-iterative/task/common/ssh"
)

func New(ctx context.Context, cloud common.Cloud, identifier common.Identifier, task common.Task) (*Task, error) {
	client, err := client.New(ctx, cloud, task.Tags)
	if err != nil {
		return nil, err
	}

	t := new(Task)
	t.Client = client
	t.Identifier = identifier
	t.Attributes = task
	t.DataSources.DefaultVPC = resources.NewDefaultVPC(
		t.Client,
	)
	t.DataSources.DefaultVPCSubnet = resources.NewDefaultVPCSubnet(
		t.Client,
		t.DataSources.DefaultVPC,
	)
	t.DataSources.Image = resources.NewImage(
		t.Client,
		t.Attributes.Environment.Image,
	)
	t.Resources.Bucket = resources.NewBucket(
		t.Client,
		t.Identifier,
	)
	t.DataSources.Credentials = resources.NewCredentials(
		t.Client,
		t.Identifier,
		t.Resources.Bucket,
	)
	t.Resources.SecurityGroup = resources.NewSecurityGroup(
		t.Client,
		t.Identifier,
		t.DataSources.DefaultVPC,
		t.Attributes.Firewall,
	)
	t.Resources.KeyPair = resources.NewKeyPair(
		t.Client,
		t.Identifier,
	)
	t.Resources.LaunchTemplate = resources.NewLaunchTemplate(
		t.Client,
		t.Identifier,
		t.Resources.SecurityGroup,
		t.DataSources.Image,
		t.Resources.KeyPair,
		t.DataSources.Credentials,
		t.Attributes,
	)
	t.Resources.AutoScalingGroup = resources.NewAutoScalingGroup(
		t.Client,
		t.Identifier,
		t.DataSources.DefaultVPCSubnet,
		t.Resources.LaunchTemplate,
		&t.Attributes.Parallelism,
		t.Attributes.Spot,
	)
	return t, nil
}

type Task struct {
	Client      *client.Client
	Identifier  common.Identifier
	Attributes  common.Task
	DataSources struct {
		*resources.DefaultVPC
		*resources.DefaultVPCSubnet
		*resources.Image
		*resources.Credentials
	}
	Resources struct {
		*resources.Bucket
		*resources.SecurityGroup
		*resources.KeyPair
		*resources.LaunchTemplate
		*resources.AutoScalingGroup
	}
}

func (t *Task) Create(ctx context.Context) error {
	logrus.Debug("Creating DefaultVPC...")
	if err := t.DataSources.DefaultVPC.Read(ctx); err != nil {
		return err
	}
	logrus.Debug("Creating DefaultVPCSubnet...")
	if err := t.DataSources.DefaultVPCSubnet.Read(ctx); err != nil {
		return err
	}
	logrus.Debug("Creating Image...")
	if err := t.DataSources.Image.Read(ctx); err != nil {
		return err
	}
	logrus.Debug("Creating Bucket...")
	if err := t.Resources.Bucket.Create(ctx); err != nil {
		return err
	}
	logrus.Debug("Creating SecurityGroup...")
	if err := t.Resources.SecurityGroup.Create(ctx); err != nil {
		return err
	}
	logrus.Debug("Creating KeyPair...")
	if err := t.Resources.KeyPair.Create(ctx); err != nil {
		return err
	}
	logrus.Debug("Creating Credentials...")
	if err := t.DataSources.Credentials.Read(ctx); err != nil {
		return err
	}
	logrus.Debug("Creating LaunchTemplate...")
	if err := t.Resources.LaunchTemplate.Create(ctx); err != nil {
		return err
	}
	logrus.Debug("Creating AutoScalingGroup...")
	if err := t.Resources.AutoScalingGroup.Create(ctx); err != nil {
		return err
	}
	logrus.Debug("Uploading Directory...")
	if t.Attributes.Environment.Directory != "" {
		if err := t.Push(ctx, t.Attributes.Environment.Directory); err != nil {
			return err
		}
	}
	logrus.Debug("Starting task...")
	if err := t.Start(ctx); err != nil {
		return err
	}
	logrus.Debug("Done!")
	t.Attributes.Addresses = t.Resources.AutoScalingGroup.Attributes.Addresses
	t.Attributes.Status = t.Resources.AutoScalingGroup.Attributes.Status
	t.Attributes.Events = t.Resources.AutoScalingGroup.Attributes.Events
	return nil
}

func (t *Task) Read(ctx context.Context) error {
	logrus.Debug("Reading DefaultVPC...")
	if err := t.DataSources.DefaultVPC.Read(ctx); err != nil {
		return err
	}
	logrus.Debug("Reading DefaultVPCSubnet...")
	if err := t.DataSources.DefaultVPCSubnet.Read(ctx); err != nil {
		return err
	}
	logrus.Debug("Reading Image...")
	if err := t.DataSources.Image.Read(ctx); err != nil {
		return err
	}
	logrus.Debug("Reading Bucket...")
	if err := t.Resources.Bucket.Read(ctx); err != nil {
		return err
	}
	logrus.Debug("Reading Credentials...")
	if err := t.DataSources.Credentials.Read(ctx); err != nil {
		return err
	}
	logrus.Debug("Reading SecurityGroup...")
	if err := t.Resources.SecurityGroup.Read(ctx); err != nil {
		return err
	}
	logrus.Debug("Reading KeyPair...")
	if err := t.Resources.KeyPair.Read(ctx); err != nil {
		return err
	}
	logrus.Debug("Reading Credentials...")
	if err := t.DataSources.Credentials.Read(ctx); err != nil {
		return err
	}
	logrus.Debug("Reading LaunchTemplate...")
	if err := t.Resources.LaunchTemplate.Read(ctx); err != nil {
		return err
	}
	logrus.Debug("Reading AutoScalingGroup...")
	if err := t.Resources.AutoScalingGroup.Read(ctx); err != nil {
		return err
	}
	logrus.Debug("Done!")
	t.Attributes.Addresses = t.Resources.AutoScalingGroup.Attributes.Addresses
	t.Attributes.Status = t.Resources.AutoScalingGroup.Attributes.Status
	t.Attributes.Events = t.Resources.AutoScalingGroup.Attributes.Events
	return nil
}

func (t *Task) Delete(ctx context.Context) error {
	logrus.Debug("Downloading Directory...")
	if t.Read(ctx) == nil {
		if t.Attributes.Environment.DirectoryOut != "" {
			if err := t.Pull(ctx, t.Attributes.Environment.Directory, t.Attributes.Environment.DirectoryOut); err != nil && err != common.NotFoundError {
				return err
			}
		}
		logrus.Debug("Emptying Bucket...")
		if err := machine.Delete(ctx, (*t.DataSources.Credentials.Resource)["RCLONE_REMOTE"]); err != nil && err != common.NotFoundError {
			return err
		}
	}
	logrus.Debug("Deleting AutoScalingGroup...")
	if err := t.Resources.AutoScalingGroup.Delete(ctx); err != nil {
		return err
	}
	logrus.Debug("Deleting LaunchTemplate...")
	if err := t.Resources.LaunchTemplate.Delete(ctx); err != nil {
		return err
	}
	logrus.Debug("Deleting KeyPair...")
	if err := t.Resources.KeyPair.Delete(ctx); err != nil {
		return err
	}
	logrus.Debug("Deleting SecurityGroup...")
	if err := t.Resources.SecurityGroup.Delete(ctx); err != nil {
		return err
	}
	logrus.Debug("Deleting Credentials...")
	if err := t.DataSources.Credentials.Read(ctx); err != nil {
		return err
	}
	logrus.Debug("Deleting Bucket...")
	if err := t.Resources.Bucket.Delete(ctx); err != nil {
		return err
	}
	logrus.Debug("Done!")
	return nil
}

func (t *Task) Logs(ctx context.Context) ([]string, error) {
	if err := t.Read(ctx); err != nil {
		return nil, err
	}

	return machine.Logs(ctx, (*t.DataSources.Credentials.Resource)["RCLONE_REMOTE"])
}

func (t *Task) Pull(ctx context.Context, destination, include string) error {
	if err := t.Read(ctx); err != nil {
		return err
	}

	return machine.Transfer(ctx, (*t.DataSources.Credentials.Resource)["RCLONE_REMOTE"]+"/data", destination, "/"+include)
}

func (t *Task) Push(ctx context.Context, source string) error {
	if err := t.Read(ctx); err != nil {
		return err
	}

	return machine.Transfer(ctx, source, (*t.DataSources.Credentials.Resource)["RCLONE_REMOTE"]+"/data", "/**")
}

func (t *Task) Start(ctx context.Context) error {
	return t.Resources.AutoScalingGroup.Update(ctx)
}

func (t *Task) Stop(ctx context.Context) error {
	original := t.Attributes.Parallelism
	defer func() { t.Attributes.Parallelism = original }()

	t.Attributes.Parallelism = 0
	return t.Resources.AutoScalingGroup.Update(ctx)
}

func (t *Task) GetAddresses(ctx context.Context) []net.IP {
	return t.Attributes.Addresses
}

func (t *Task) Events(ctx context.Context) []common.Event {
	return t.Attributes.Events
}

func (t *Task) Status(ctx context.Context) (common.Status, error) {
	if err := t.Read(ctx); err != nil {
		return nil, err
	}

	return machine.Status(ctx, (*t.DataSources.Credentials.Resource)["RCLONE_REMOTE"], t.Attributes.Status)
}

func (t *Task) GetKeyPair(ctx context.Context) (*ssh.DeterministicSSHKeyPair, error) {
	return t.Client.GetKeyPair(ctx)
}

func (t *Task) GetIdentifier(ctx context.Context) common.Identifier {
	return t.Identifier
}
