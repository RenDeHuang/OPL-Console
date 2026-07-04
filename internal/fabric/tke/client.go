package tke

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/RenDeHuang/opl-console/internal/fabric"
)

const (
	workspaceContainerName  = "workspace"
	workspaceHTTPPort       = int32(3000)
	defaultStorageMountPath = "/data"
	maxDNS1123LabelLength   = 63
	defaultConsoleService   = "opl-console"
	defaultConsolePort      = int32(8787)

	originalComputeIDAnnotation   = "opl.one-person-lab/original-compute-id"
	originalStorageIDAnnotation   = "opl.one-person-lab/original-storage-id"
	originalWorkspaceIDAnnotation = "opl.one-person-lab/original-workspace-id"
)

type Config struct {
	Namespace    string
	Image        string
	StorageClass string
	IngressClass string

	ConsoleServiceName string
	ConsoleServicePort int32
}

type Client struct {
	cfg    Config
	client kubernetes.Interface
}

var _ fabric.Port = (*Client)(nil)

func New(cfg Config, client kubernetes.Interface) *Client {
	return &Client{cfg: cfg, client: client}
}

func (c *Client) consoleServiceName() string {
	if c.cfg.ConsoleServiceName != "" {
		return c.cfg.ConsoleServiceName
	}
	return defaultConsoleService
}

func (c *Client) consoleServicePort() int32 {
	if c.cfg.ConsoleServicePort != 0 {
		return c.cfg.ConsoleServicePort
	}
	return defaultConsolePort
}

func (c *Client) CreateCompute(ctx context.Context, request fabric.CreateComputeRequest) (fabric.RuntimeHandle, error) {
	name := computeName(request.ComputeID)
	labels := computeLabels(request.ComputeID)
	replicas := int32(1)

	if _, err := c.client.AppsV1().Deployments(c.cfg.Namespace).Create(ctx, &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Labels:      labels,
			Annotations: originalComputeAnnotations(request.ComputeID),
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{MatchLabels: labels},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: labels},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name:  workspaceContainerName,
						Image: c.cfg.Image,
						Ports: []corev1.ContainerPort{{ContainerPort: workspaceHTTPPort}},
					}},
				},
			},
		},
	}, metav1.CreateOptions{}); err != nil {
		return fabric.RuntimeHandle{}, fmt.Errorf("create deployment: %w", err)
	}

	if _, err := c.client.CoreV1().Services(c.cfg.Namespace).Create(ctx, &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Labels:      labels,
			Annotations: originalComputeAnnotations(request.ComputeID),
		},
		Spec: corev1.ServiceSpec{
			Type:     corev1.ServiceTypeClusterIP,
			Selector: labels,
			Ports: []corev1.ServicePort{{
				Name: "http",
				Port: workspaceHTTPPort,
			}},
		},
	}, metav1.CreateOptions{}); err != nil {
		rollbackErr := ignoreNotFound(c.client.AppsV1().Deployments(c.cfg.Namespace).Delete(ctx, name, metav1.DeleteOptions{}))
		if rollbackErr != nil {
			return fabric.RuntimeHandle{}, errors.Join(
				fmt.Errorf("create service: %w", err),
				fmt.Errorf("rollback deployment: %w", rollbackErr),
			)
		}
		return fabric.RuntimeHandle{}, fmt.Errorf("create service: %w", err)
	}

	return fabric.RuntimeHandle{
		ProviderResourceID: "deployment/" + name,
		Status:             "running",
	}, nil
}

func (c *Client) CreateStorage(ctx context.Context, request fabric.CreateStorageRequest) (fabric.RuntimeHandle, error) {
	name := storageName(request.StorageID)
	var storageClassName *string
	if c.cfg.StorageClass != "" {
		storageClass := c.cfg.StorageClass
		storageClassName = &storageClass
	}
	size := fmt.Sprintf("%dGi", request.Package.StorageGB)
	if request.Package.StorageGB <= 0 {
		size = "10Gi"
	}

	if _, err := c.client.CoreV1().PersistentVolumeClaims(c.cfg.Namespace).Create(ctx, &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Labels:      storageLabels(request.StorageID),
			Annotations: originalStorageAnnotations(request.StorageID),
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes:      []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
			StorageClassName: storageClassName,
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse(size),
				},
			},
		},
	}, metav1.CreateOptions{}); err != nil {
		return fabric.RuntimeHandle{}, fmt.Errorf("create pvc: %w", err)
	}

	return fabric.RuntimeHandle{
		ProviderResourceID: "pvc/" + name,
		Status:             "available",
	}, nil
}

func (c *Client) StopCompute(ctx context.Context, request fabric.StopComputeRequest) (fabric.RuntimeHandle, error) {
	deployments := c.client.AppsV1().Deployments(c.cfg.Namespace)
	name := computeName(request.ComputeID)
	deployment, err := deployments.Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return fabric.RuntimeHandle{}, fmt.Errorf("get deployment %q: %w", name, err)
	}
	replicas := int32(0)
	deployment.Spec.Replicas = &replicas
	if _, err := deployments.Update(ctx, deployment, metav1.UpdateOptions{}); err != nil {
		return fabric.RuntimeHandle{}, fmt.Errorf("stop deployment %q: %w", name, err)
	}
	return fabric.RuntimeHandle{ProviderResourceID: "deployment/" + name, Status: "stopped"}, nil
}

func (c *Client) RestartCompute(ctx context.Context, request fabric.RestartComputeRequest) (fabric.RuntimeHandle, error) {
	deployments := c.client.AppsV1().Deployments(c.cfg.Namespace)
	name := computeName(request.ComputeID)
	deployment, err := deployments.Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return fabric.RuntimeHandle{}, fmt.Errorf("get deployment %q: %w", name, err)
	}
	replicas := int32(1)
	deployment.Spec.Replicas = &replicas
	if _, err := deployments.Update(ctx, deployment, metav1.UpdateOptions{}); err != nil {
		return fabric.RuntimeHandle{}, fmt.Errorf("restart deployment %q: %w", name, err)
	}
	return fabric.RuntimeHandle{ProviderResourceID: "deployment/" + name, Status: "running"}, nil
}

func (c *Client) AttachStorage(ctx context.Context, request fabric.AttachStorageRequest) (fabric.RuntimeHandle, error) {
	deployments := c.client.AppsV1().Deployments(c.cfg.Namespace)
	deploymentName := computeName(request.ComputeID)
	deployment, err := deployments.Get(ctx, deploymentName, metav1.GetOptions{})
	if err != nil {
		return fabric.RuntimeHandle{}, fmt.Errorf("get deployment %q: %w", deploymentName, err)
	}

	workspaceIndex := -1
	for i, container := range deployment.Spec.Template.Spec.Containers {
		if container.Name == workspaceContainerName {
			workspaceIndex = i
			break
		}
	}
	if workspaceIndex == -1 {
		return fabric.RuntimeHandle{}, fmt.Errorf("workspace container %q not found in deployment %q", workspaceContainerName, deploymentName)
	}

	volumeName := storageName(request.StorageID)
	volumeExists := false
	for _, volume := range deployment.Spec.Template.Spec.Volumes {
		if volume.Name == volumeName {
			volumeExists = true
			break
		}
	}
	if !volumeExists {
		deployment.Spec.Template.Spec.Volumes = append(deployment.Spec.Template.Spec.Volumes, corev1.Volume{
			Name: volumeName,
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: volumeName,
				},
			},
		})
	}

	mountPath := request.MountPath
	if mountPath == "" {
		mountPath = defaultStorageMountPath
	}
	mounts := &deployment.Spec.Template.Spec.Containers[workspaceIndex].VolumeMounts
	mountExists := false
	for _, mount := range *mounts {
		if mount.Name == volumeName {
			mountExists = true
			break
		}
	}
	if !mountExists {
		*mounts = append(*mounts, corev1.VolumeMount{
			Name:      volumeName,
			MountPath: mountPath,
		})
	}

	if _, err := deployments.Update(ctx, deployment, metav1.UpdateOptions{}); err != nil {
		return fabric.RuntimeHandle{}, fmt.Errorf("update deployment %q: %w", deploymentName, err)
	}

	return fabric.RuntimeHandle{
		ProviderResourceID: request.AttachmentID,
		Status:             "attached",
	}, nil
}

func (c *Client) DetachStorage(ctx context.Context, request fabric.DetachStorageRequest) (fabric.RuntimeHandle, error) {
	deployments := c.client.AppsV1().Deployments(c.cfg.Namespace)
	deploymentName := computeName(request.ComputeID)
	deployment, err := deployments.Get(ctx, deploymentName, metav1.GetOptions{})
	if err != nil {
		return fabric.RuntimeHandle{}, fmt.Errorf("get deployment %q: %w", deploymentName, err)
	}
	volumeName := storageName(request.StorageID)
	for i := range deployment.Spec.Template.Spec.Containers {
		if deployment.Spec.Template.Spec.Containers[i].Name != workspaceContainerName {
			continue
		}
		filtered := deployment.Spec.Template.Spec.Containers[i].VolumeMounts[:0]
		for _, mount := range deployment.Spec.Template.Spec.Containers[i].VolumeMounts {
			if mount.Name != volumeName {
				filtered = append(filtered, mount)
			}
		}
		deployment.Spec.Template.Spec.Containers[i].VolumeMounts = filtered
	}
	filteredVolumes := deployment.Spec.Template.Spec.Volumes[:0]
	for _, volume := range deployment.Spec.Template.Spec.Volumes {
		if volume.Name != volumeName {
			filteredVolumes = append(filteredVolumes, volume)
		}
	}
	deployment.Spec.Template.Spec.Volumes = filteredVolumes
	if _, err := deployments.Update(ctx, deployment, metav1.UpdateOptions{}); err != nil {
		return fabric.RuntimeHandle{}, fmt.Errorf("detach storage from deployment %q: %w", deploymentName, err)
	}
	return fabric.RuntimeHandle{ProviderResourceID: request.AttachmentID, Status: "detached_retained"}, nil
}

func (c *Client) CreateWorkspaceRoute(ctx context.Context, request fabric.CreateRouteRequest) (fabric.RuntimeHandle, error) {
	// v1 deliberately routes /w traffic to the Console validator before workspace handoff.
	// The Kubernetes Secret is runtime handoff state, not ingress auth enforcement.
	tokenUpsert, err := c.upsertTokenSecret(ctx, request.WorkspaceID, request.ComputeID, request.Token)
	if err != nil {
		return fabric.RuntimeHandle{}, fmt.Errorf("upsert token secret: %w", err)
	}

	name := routeName(request.WorkspaceID)
	pathType := networkingv1.PathTypePrefix
	ingress := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Labels:      routeLabels(request.WorkspaceID),
			Annotations: routeAnnotations(request.WorkspaceID, request.ComputeID),
		},
		Spec: networkingv1.IngressSpec{
			Rules: []networkingv1.IngressRule{{
				IngressRuleValue: networkingv1.IngressRuleValue{
					HTTP: &networkingv1.HTTPIngressRuleValue{
						Paths: []networkingv1.HTTPIngressPath{{
							Path:     workspacePath(request.WorkspaceID),
							PathType: &pathType,
							Backend: networkingv1.IngressBackend{
								Service: &networkingv1.IngressServiceBackend{
									Name: c.consoleServiceName(),
									Port: networkingv1.ServiceBackendPort{Number: c.consoleServicePort()},
								},
							},
						}},
					},
				},
			}},
		},
	}
	if c.cfg.IngressClass != "" {
		ingress.Spec.IngressClassName = &c.cfg.IngressClass
	}

	if _, err := c.client.NetworkingV1().Ingresses(c.cfg.Namespace).Create(ctx, ingress, metav1.CreateOptions{}); err != nil {
		rollbackErr := c.rollbackTokenSecret(ctx, tokenUpsert)
		if rollbackErr != nil {
			return fabric.RuntimeHandle{}, errors.Join(
				fmt.Errorf("create ingress: %w", err),
				fmt.Errorf("rollback token secret: %w", rollbackErr),
			)
		}
		return fabric.RuntimeHandle{}, fmt.Errorf("create ingress: %w", err)
	}

	return routeHandle(request.WorkspaceID, request.Token), nil
}

func (c *Client) CreateStorageBackup(ctx context.Context, request fabric.BackupStorageRequest) (fabric.RuntimeHandle, error) {
	return fabric.RuntimeHandle{
		ProviderResourceID: "backup/" + dns1123Name(request.BackupID),
		Status:             "ready",
	}, nil
}

func (c *Client) RestoreStorageBackup(ctx context.Context, request fabric.RestoreStorageRequest) (fabric.RuntimeHandle, error) {
	return fabric.RuntimeHandle{
		ProviderResourceID: "backup/" + dns1123Name(request.BackupID),
		Status:             "restored",
	}, nil
}

func (c *Client) DestroyCompute(ctx context.Context, request fabric.DestroyComputeRequest) error {
	deployments := c.client.AppsV1().Deployments(c.cfg.Namespace)
	services := c.client.CoreV1().Services(c.cfg.Namespace)
	name := computeName(request.ComputeID)

	return errors.Join(
		ignoreNotFound(deployments.Delete(ctx, name, metav1.DeleteOptions{})),
		ignoreNotFound(services.Delete(ctx, name, metav1.DeleteOptions{})),
	)
}

func (c *Client) DestroyStorage(ctx context.Context, request fabric.DestroyStorageRequest) error {
	return ignoreNotFound(c.client.CoreV1().PersistentVolumeClaims(c.cfg.Namespace).Delete(ctx, storageName(request.StorageID), metav1.DeleteOptions{}))
}

func (c *Client) DestroyWorkspaceRoute(ctx context.Context, request fabric.DestroyWorkspaceRouteRequest) error {
	return errors.Join(
		ignoreNotFound(c.client.NetworkingV1().Ingresses(c.cfg.Namespace).Delete(ctx, routeName(request.WorkspaceID), metav1.DeleteOptions{})),
		c.DeleteWorkspaceToken(ctx, fabric.DeleteWorkspaceTokenRequest{WorkspaceID: request.WorkspaceID}),
	)
}

func (c *Client) ResetWorkspaceToken(ctx context.Context, request fabric.ResetWorkspaceTokenRequest) (fabric.RuntimeHandle, error) {
	if _, err := c.upsertTokenSecret(ctx, request.WorkspaceID, "", request.Token); err != nil {
		return fabric.RuntimeHandle{}, fmt.Errorf("upsert token secret: %w", err)
	}
	return routeHandle(request.WorkspaceID, request.Token), nil
}

func (c *Client) DeleteWorkspaceToken(ctx context.Context, request fabric.DeleteWorkspaceTokenRequest) error {
	return ignoreNotFound(c.client.CoreV1().Secrets(c.cfg.Namespace).Delete(ctx, tokenSecretName(request.WorkspaceID), metav1.DeleteOptions{}))
}

func (c *Client) RuntimeStatus(ctx context.Context, request fabric.RuntimeStatusRequest) (fabric.RuntimeStatus, error) {
	status := fabric.RuntimeStatus{
		WorkspaceID:  request.WorkspaceID,
		ComputeState: "unknown",
		StorageState: "unknown",
		RouteState:   "unknown",
		Metadata:     map[string]string{"namespace": c.cfg.Namespace},
	}
	if request.ComputeID != "" {
		deployment, err := c.client.AppsV1().Deployments(c.cfg.Namespace).Get(ctx, computeName(request.ComputeID), metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			status.ComputeState = "not_found"
		} else if err != nil {
			return status, fmt.Errorf("get compute status: %w", err)
		} else if deployment.Spec.Replicas != nil && *deployment.Spec.Replicas == 0 {
			status.ComputeState = "stopped"
		} else if deployment.Status.ReadyReplicas > 0 {
			status.ComputeState = "running"
		} else {
			status.ComputeState = "provisioning"
		}
	}
	if request.StorageID != "" {
		_, err := c.client.CoreV1().PersistentVolumeClaims(c.cfg.Namespace).Get(ctx, storageName(request.StorageID), metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			status.StorageState = "not_found"
		} else if err != nil {
			return status, fmt.Errorf("get storage status: %w", err)
		} else {
			status.StorageState = "available"
		}
	}
	_, err := c.client.NetworkingV1().Ingresses(c.cfg.Namespace).Get(ctx, routeName(request.WorkspaceID), metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		status.RouteState = "not_found"
	} else if err != nil {
		return status, fmt.Errorf("get route status: %w", err)
	} else {
		status.RouteState = "ready"
	}
	status.Ready = status.ComputeState == "running" && status.StorageState != "not_found" && status.RouteState == "ready"
	return status, nil
}

type tokenSecretUpsert struct {
	name     string
	created  bool
	previous *corev1.Secret
}

func (c *Client) upsertTokenSecret(ctx context.Context, workspaceID, computeID, token string) (tokenSecretUpsert, error) {
	secrets := c.client.CoreV1().Secrets(c.cfg.Namespace)
	name := tokenSecretName(workspaceID)

	secret, err := secrets.Get(ctx, name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		if _, err := secrets.Create(ctx, tokenSecret(workspaceID, computeID, token), metav1.CreateOptions{}); err != nil {
			return tokenSecretUpsert{}, err
		}
		return tokenSecretUpsert{name: name, created: true}, nil
	}
	if err != nil {
		return tokenSecretUpsert{}, err
	}

	previous := secret.DeepCopy()
	secret.Data = map[string][]byte{"token": []byte(token)}
	secret.Labels = mergeStringMap(secret.Labels, routeLabels(workspaceID))
	secret.Annotations = mergeStringMap(secret.Annotations, routeAnnotations(workspaceID, computeID))
	if _, err := secrets.Update(ctx, secret, metav1.UpdateOptions{}); err != nil {
		return tokenSecretUpsert{}, err
	}
	return tokenSecretUpsert{name: name, previous: previous}, nil
}

func (c *Client) rollbackTokenSecret(ctx context.Context, upsert tokenSecretUpsert) error {
	secrets := c.client.CoreV1().Secrets(c.cfg.Namespace)
	if upsert.created {
		return ignoreNotFound(secrets.Delete(ctx, upsert.name, metav1.DeleteOptions{}))
	}
	if upsert.previous == nil {
		return nil
	}
	current, err := secrets.Get(ctx, upsert.name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		restored := upsert.previous.DeepCopy()
		restored.ResourceVersion = ""
		_, err = secrets.Create(ctx, restored, metav1.CreateOptions{})
		return err
	}
	if err != nil {
		return err
	}
	restored := upsert.previous.DeepCopy()
	restored.ResourceVersion = current.ResourceVersion
	_, err = secrets.Update(ctx, restored, metav1.UpdateOptions{})
	return err
}

func tokenSecret(workspaceID, computeID, token string) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:        tokenSecretName(workspaceID),
			Labels:      routeLabels(workspaceID),
			Annotations: routeAnnotations(workspaceID, computeID),
		},
		Type: corev1.SecretTypeOpaque,
		Data: map[string][]byte{"token": []byte(token)},
	}
}

func routeHandle(workspaceID, token string) fabric.RuntimeHandle {
	return fabric.RuntimeHandle{
		ProviderResourceID: "ingress/" + routeName(workspaceID),
		Status:             "ready",
		URL:                workspacePath(workspaceID) + "?token=" + token,
	}
}

func ignoreNotFound(err error) error {
	if apierrors.IsNotFound(err) {
		return nil
	}
	return err
}

func computeLabels(computeID string) map[string]string {
	return map[string]string{
		"app.kubernetes.io/name":      "opl-workspace",
		"app.kubernetes.io/component": "compute",
		"opl.one-person-lab/compute":  computeName(computeID),
	}
}

func storageLabels(storageID string) map[string]string {
	return map[string]string{
		"app.kubernetes.io/name":      "opl-workspace",
		"app.kubernetes.io/component": "storage",
		"opl.one-person-lab/storage":  storageName(storageID),
	}
}

func routeLabels(workspaceID string) map[string]string {
	return map[string]string{
		"app.kubernetes.io/name":       "opl-workspace",
		"app.kubernetes.io/component":  "route",
		"opl.one-person-lab/workspace": routeName(workspaceID),
	}
}

func originalComputeAnnotations(computeID string) map[string]string {
	return map[string]string{originalComputeIDAnnotation: computeID}
}

func originalStorageAnnotations(storageID string) map[string]string {
	return map[string]string{originalStorageIDAnnotation: storageID}
}

func originalWorkspaceAnnotations(workspaceID string) map[string]string {
	return map[string]string{originalWorkspaceIDAnnotation: workspaceID}
}

func routeAnnotations(workspaceID, computeID string) map[string]string {
	annotations := originalWorkspaceAnnotations(workspaceID)
	if computeID != "" {
		annotations[originalComputeIDAnnotation] = computeID
	}
	return annotations
}

func computeName(computeID string) string {
	return dns1123Name(computeID)
}

func storageName(storageID string) string {
	return dns1123Name(storageID)
}

func tokenSecretName(workspaceID string) string {
	return dns1123Name("workspace-" + workspaceID + "-token")
}

func routeName(workspaceID string) string {
	return dns1123Name(workspaceID)
}

func workspacePath(workspaceID string) string {
	return "/w/" + workspaceID
}

func dns1123Name(value string) string {
	hash := shortHash(value)
	name := sanitizeDNS1123(value)
	if name == "" {
		name = "x-" + hash
	}
	if name == value && len(name) <= maxDNS1123LabelLength {
		return name
	}

	prefixLength := maxDNS1123LabelLength - len(hash) - 1
	if len(name) < prefixLength {
		prefixLength = len(name)
	}
	prefix := strings.Trim(name[:prefixLength], "-")
	if prefix == "" {
		return "x-" + hash
	}
	return prefix + "-" + hash
}

func sanitizeDNS1123(value string) string {
	var b strings.Builder
	lastDash := false
	for _, r := range strings.ToLower(value) {
		valid := (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9')
		if valid {
			b.WriteRune(r)
			lastDash = false
			continue
		}
		if !lastDash {
			b.WriteByte('-')
			lastDash = true
		}
	}
	return strings.Trim(b.String(), "-")
}

func shortHash(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])[:8]
}

func mergeStringMap(existing, required map[string]string) map[string]string {
	if existing == nil {
		existing = map[string]string{}
	}
	for key, value := range required {
		existing[key] = value
	}
	return existing
}
