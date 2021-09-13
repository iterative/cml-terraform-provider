package kubernetes

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"time"

	terraform_resource "github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	terraform_schema "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	kubernetes_batch "k8s.io/api/batch/v1"
	kubernetes_core "k8s.io/api/core/v1"
	kubernetes_errors "k8s.io/apimachinery/pkg/api/errors"
	kubernetes_resource "k8s.io/apimachinery/pkg/api/resource"
	kubernetes_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubernetes_wait "k8s.io/apimachinery/pkg/util/wait"
	kubernetes "k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	kubernetes_clientcmd "k8s.io/client-go/tools/clientcmd"
)

// Create a "machine" (actually a Kubernetes job) on the cluster.
func ResourceMachineCreate(ctx context.Context, d *terraform_schema.ResourceData, meta interface{}) error {
	conn, namespace, err := kubernetesClient()
	if err != nil {
		return err
	}

	// Get the instance resource specification from its vendor-agnostic name.
	instanceType, err := getInstanceType(d.Get("instance_type").(string), d.Get("instance_gpu").(string))
	if err != nil {
		return err
	}

	// Define the job namespace and name.
	jobName := d.Id()
	jobNamespace := namespace

	// Define the metadata
	jobMetadata := map[string]string{}
	for key, value := range d.Get("metadata").(map[string]interface{}) {
		jobMetadata[key] = value.(string)
	}

	// Define the accelerator settings (i.e. GPU type, model, ...)
	jobNodeSelector := map[string]string{}
	jobAccelerator := instanceType["accelerator"]["model"]
	jobGPUType := instanceType["accelerator"]["type"]
	jobGPUCount := instanceType["accelerator"]["count"]

	// Define the dynamic resource allocation limits for the job pods.
	jobLimits := kubernetes_core.ResourceList{}
	jobLimits[kubernetes_core.ResourceName("memory")] = kubernetes_resource.MustParse(instanceType["memory"]["amount"])
	jobLimits[kubernetes_core.ResourceName("cpu")] = kubernetes_resource.MustParse(instanceType["cores"]["count"])
	if diskAmount := d.Get("instance_hdd_size").(int); diskAmount > 0 {
		jobLimits[kubernetes_core.ResourceName("ephemeral-storage")] = kubernetes_resource.MustParse(strconv.Itoa(diskAmount) + "G")
	}

	// Use the default CML Docker image unless specified otherwise.
	jobImageName := jobName
	jobImage := d.Get("image").(string)
	if jobImage == "" {
		jobImage = "dvcorg/cml:0-dvc2-base1-gpu"
	}

	// Script to run on the container instead of the default entry point.
	jobStartupScript, err := base64.StdEncoding.DecodeString(d.Get("startup_script").(string))
	if err != nil {
		return err
	}

	// If the resource requires GPU provisioning, determine how many GPUs and the kind of GPU it needs.
	if jobGPUCount > "0" {
		jobLimits[kubernetes_core.ResourceName(jobGPUType)] = kubernetes_resource.MustParse(jobGPUCount)
		if jobAccelerator != "" {
			jobNodeSelector = map[string]string{"accelerator": jobAccelerator}
		}
	}

	// Leave the job running for 30 seconds after the termination signal, but remove it immediately after terminating.
	jobTTLSecondsAfterFinished := int32(0)
	jobTerminationGracePeriod := int64(30)

	// Don't run many pods in parallel.
	jobBackoffLimit := int32(1)
	jobCompletions := int32(1)
	jobParallelism := int32(1)

	job := kubernetes_batch.Job{
		ObjectMeta: kubernetes_meta.ObjectMeta{
			Name:        jobName,
			Namespace:   jobNamespace,
			Labels:      jobMetadata,
			Annotations: jobMetadata,
		},
		Spec: kubernetes_batch.JobSpec{
			BackoffLimit:            &jobBackoffLimit,
			Completions:             &jobCompletions,
			Parallelism:             &jobParallelism,
			TTLSecondsAfterFinished: &jobTTLSecondsAfterFinished,
			Template: kubernetes_core.PodTemplateSpec{
				Spec: kubernetes_core.PodSpec{
					TerminationGracePeriodSeconds: &jobTerminationGracePeriod,
					RestartPolicy:                 "Never", // We don't want pods to restart on failure.
					NodeSelector:                  jobNodeSelector,
					Containers: []kubernetes_core.Container{
						{
							Name:  jobImageName,
							Image: jobImage,
							Resources: kubernetes_core.ResourceRequirements{
								Limits: jobLimits,
								Requests: kubernetes_core.ResourceList{
									// Don't allocate any resources statically and let the pod scale vertically when and if required.
									kubernetes_core.ResourceName("memory"): kubernetes_resource.MustParse("0"),
									kubernetes_core.ResourceName("cpu"):    kubernetes_resource.MustParse("0"),
								},
							},
							Command: []string{
								"bash", "-c", string(jobStartupScript),
							},
						},
					},
				},
			},
		},
	}

	// Ask Kubernetes to create the job.
	log.Printf("[INFO] Creating new Job: %#v", job)
	out, err := conn.BatchV1().Jobs(jobNamespace).Create(ctx, &job, kubernetes_meta.CreateOptions{})
	if err != nil {
		return fmt.Errorf("Failed to create Job! API error: %s", err)
	}
	log.Printf("[INFO] Submitted new job: %#v", out)

	// Get the controller unique identifier for the job, so we can easily find the pods it creates.
	waitSelector := fmt.Sprintf("controller-uid=%s", out.GetObjectMeta().GetLabels()["controller-uid"])

	// Wait for the job to satisfy the readiness condition specified through kubernetes_readiness_command.
	return terraform_resource.Retry(d.Timeout(terraform_schema.TimeoutCreate), func() *terraform_resource.RetryError {
		// If the pod awaiting function fails because the pod doesn't exist yet, tell Terraform to keep trying.
		if err := waitForPods(conn, 1*time.Second, 1*time.Second, jobNamespace, waitSelector); err != nil {
			return terraform_resource.RetryableError(fmt.Errorf("Still creating... %s", err))
			// Else, if the awaiting function detects pods for the job but aren't ready yet.
		} else if err := ResourceMachineCheck(ctx, d, meta); err != nil {
			return terraform_resource.NonRetryableError(err)
			// Else, job pods are running and ready.
		} else {
			return nil
		}
	})
}

// Delete a "machine" (actually a Kubernetes job) from the cluster.
func ResourceMachineDelete(ctx context.Context, d *terraform_schema.ResourceData, meta interface{}) error {
	conn, namespace, err := kubernetesClient()
	if err != nil {
		return err
	}

	log.Printf("[INFO] Deleting job: %#v", d.Id())
	_, err = conn.BatchV1().Jobs(namespace).Get(ctx, d.Id(), kubernetes_meta.GetOptions{})
	if err != nil {
		if statusErr, ok := err.(*kubernetes_errors.StatusError); ok && statusErr.ErrStatus.Code == 404 {
			log.Printf("[INFO] Job %#v doesn't exist; skipping deletion...", d.Id())
			return nil
		}
		log.Printf("[DEBUG] Received error: %#v", err)
		return err
	}

	// DeletePropagationForeground deletes the resources and causes the garbage
	// collector to delete dependent resources and wait for all dependents whose
	// ownerReference.blockOwnerDeletion=true.
	propagationPolicy := kubernetes_meta.DeletePropagationForeground

	err = conn.BatchV1().Jobs(namespace).Delete(ctx, d.Id(), kubernetes_meta.DeleteOptions{
		PropagationPolicy: &propagationPolicy,
	})
	if err != nil {
		return fmt.Errorf("Failed to delete Job! API error: %s", err)
	}

	err = terraform_resource.RetryContext(ctx, d.Timeout(terraform_schema.TimeoutDelete), func() *terraform_resource.RetryError {
		_, err := conn.BatchV1().Jobs(namespace).Get(ctx, d.Id(), kubernetes_meta.GetOptions{})
		if err != nil {
			if statusErr, ok := err.(*kubernetes_errors.StatusError); ok && statusErr.ErrStatus.Code == 404 {
				return nil
			}
			return terraform_resource.NonRetryableError(err)
		}

		e := fmt.Errorf("Job %s still exists", d.Id())
		return terraform_resource.RetryableError(e)
	})
	if err != nil {
		return err
	}

	log.Printf("[INFO] Job %s deleted", d.Id())

	d.SetId("")
	return nil
}

// Check if a "machine" (actually a Kubernetes job) exists on the cluster.
func ResourceMachineCheck(ctx context.Context, d *terraform_schema.ResourceData, meta interface{}) error {
	conn, namespace, err := kubernetesClient()
	if err != nil {
		return err
	}

	log.Printf("[INFO] Checking job %s", d.Id())
	_, err = conn.BatchV1().Jobs(namespace).Get(ctx, d.Id(), kubernetes_meta.GetOptions{})
	if err != nil {
		if statusErr, ok := err.(*kubernetes_errors.StatusError); ok && statusErr.ErrStatus.Code == 404 {
			return nil
		}
		log.Printf("[DEBUG] Received error: %#v", err)
	}
	return err
}

// Return a Kubernetes clientset object and the default namespace from the contents of the KUBERNETES_CONFIGURATION environment variable.
func kubernetesClient() (kubernetes.Interface, string, error) {
	configurationBytes := []byte(os.Getenv("KUBERNETES_CONFIGURATION"))
	configuration, err := kubernetes_clientcmd.NewClientConfigFromBytes(configurationBytes)
	if err != nil {
		return nil, "", err
	}

	clientConfiguration, err := configuration.ClientConfig()
	if err != nil {
		return nil, "", err
	}

	client, err := kubernetes.NewForConfig(clientConfiguration)
	if err != nil {
		return nil, "", err
	}

	namespace, _, err := configuration.Namespace()
	if err != nil {
		return nil, "", err
	}

	return client, namespace, nil
}

// Wait for the pods matching the specified selector to pass their respective readiness probes.
func waitForPods(client kubernetes.Interface, interval time.Duration, timeout time.Duration, namespace string, selector string) error {
	// Retrieve all the pods matching the given selector.
	pods, err := client.
		CoreV1().Pods(namespace).
		List(context.TODO(), kubernetes_meta.ListOptions{LabelSelector: selector})

	if err != nil {
		return err
	} else if len(pods.Items) == 0 {
		return fmt.Errorf("no pods in %s matching %s", namespace, selector)
	}

	// Await the matching pods sequentially and return as soon as one of it fails the readiness check.
	for _, currentPod := range pods.Items {
		// Define a closured function in charge of checking the state of the current pod.
		function := func() (bool, error) {
			pod, err := client.
				CoreV1().
				Pods(namespace).
				Get(context.TODO(), currentPod.Name, kubernetes_meta.GetOptions{})

			if err != nil {
				return false, err
			}
			fmt.Printf("%s\n", selector)

			for _, condition := range pod.Status.Conditions {
				if condition.Type == "Ready" {
					return condition.Status == "True", nil
				}
			}
			return false, nil
		}
		// Wait for the pod to be ready by polling its status, and return the error on timeout.
		if err := kubernetes_wait.PollImmediate(interval, timeout, function); err != nil {
			return err
		}
	}
	return nil
}

// Get the actual instance characteristics from its vendor-agnostic reference.
func getInstanceType(instanceType string, instanceGPU string) (map[string]map[string]string, error) {
	// TODO: consider the use of an enumeration instead of a nested string map.
	instanceTypes := make(map[string]map[string]map[string]string)
	// Resource amounts mirror their AWS counterparts.
	instanceTypes["m"] = map[string]map[string]string{
		"accelerator": {
			"count": "0",
			"type":  "",
			"model": "",
		},
		"cores":  {"count": "8"},
		"memory": {"amount": "32Gi"},
	}
	instanceTypes["l"] = map[string]map[string]string{
		"accelerator": {
			"count": "0",
			"type":  "",
			"model": "",
		},
		"cores":  {"count": "32"},
		"memory": {"amount": "128Gi"},
	}
	instanceTypes["xl"] = map[string]map[string]string{
		"accelerator": {
			"count": "0",
			"type":  "",
			"model": "",
		},
		"cores":  {"count": "64"},
		"memory": {"amount": "256Gi"},
	}
	instanceTypes["mk80"] = map[string]map[string]string{
		"accelerator": {
			"count": "1",
			"type":  "nvidia.com/gpu",
			"model": "nvidia-tesla-k80",
		},
		"cores":  {"count": "4"},
		"memory": {"amount": "64Gi"},
	}
	instanceTypes["lk80"] = map[string]map[string]string{
		"accelerator": {
			"count": "8",
			"type":  "nvidia.com/gpu",
			"model": "nvidia-tesla-k80",
		},
		"cores":  {"count": "32"},
		"memory": {"amount": "512Gi"},
	}
	instanceTypes["xlk80"] = map[string]map[string]string{
		"accelerator": {
			"count": "16",
			"type":  "nvidia.com/gpu",
			"model": "nvidia-tesla-k80",
		},
		"cores":  {"count": "64"},
		"memory": {"amount": "768Gi"},
	}
	instanceTypes["mv100"] = map[string]map[string]string{
		"accelerator": {
			"count": "1",
			"type":  "nvidia.com/gpu",
			"model": "nvidia-tesla-v100",
		},
		"cores":  {"count": "8"},
		"memory": {"amount": "64Gi"},
	}
	instanceTypes["lv100"] = map[string]map[string]string{
		"accelerator": {
			"count": "4",
			"type":  "nvidia.com/gpu",
			"model": "nvidia-tesla-v100",
		},
		"cores":  {"count": "32"},
		"memory": {"amount": "256Gi"},
	}
	instanceTypes["xlv100"] = map[string]map[string]string{
		"accelerator": {
			"count": "8",
			"type":  "nvidia.com/gpu",
			"model": "nvidia-tesla-v100",
		},
		"cores":  {"count": "64"},
		"memory": {"amount": "512Gi"},
	}

	if val, ok := instanceTypes[instanceType+instanceGPU]; ok {
		return val, nil
	} else if val, ok := instanceTypes[instanceType]; ok {
		// Allow users to specify custom accelerator selectors.
		return map[string]map[string]string{
			"accelerator": {
				"count": val["accelerator"]["count"],
				"type":  val["accelerator"]["type"],
				"model": instanceGPU,
			},
			"cores":  val["cores"],
			"memory": val["memory"],
		}, nil
	}

	return nil, fmt.Errorf("invalid instance type")
}

func ResourceMachineLogs(ctx context.Context, d *terraform_schema.ResourceData, m interface{}) (string, error) {
	conn, namespace, err := kubernetesClient()
	if err != nil {
		return "", err
	}

	job, err := conn.BatchV1().Jobs(namespace).Get(ctx, d.Id(), kubernetes_meta.GetOptions{})
	if err != nil {
		return "", err
	}

	pods, err := conn.CoreV1().Pods(namespace).List(ctx, kubernetes_meta.ListOptions{
		LabelSelector: fmt.Sprintf("controller-uid=%s", job.GetObjectMeta().GetLabels()["controller-uid"]),
	})
	if err != nil {
		return "", err
	}

	logs, err := conn.CoreV1().Pods(namespace).GetLogs(pods.Items[0].Name, &kubernetes_core.PodLogOptions{}).Stream(ctx)
	if err != nil {
		return "", err
	}
	defer logs.Close()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, logs)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}
