package tke

import (
	"context"
	"errors"
	"fmt"

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
	workspaceContainerName = "workspace"
	workspaceHTTPPort      = int32(3000)
)

type Config struct {
	Namespace    string
	Image        string
	StorageClass string
	IngressClass string
}

type Client struct {
	cfg    Config
	client kubernetes.Interface
}

var _ fabric.Port = (*Client)(nil)

func New(cfg Config, client kubernetes.Interface) *Client {
	return &Client{cfg: cfg, client: client}
}

func (c *Client) CreateCompute(ctx context.Context, request fabric.CreateComputeRequest) (fabric.RuntimeHandle, error) {
	labels := computeLabels(request.ComputeID)
	replicas := int32(1)

	if _, err := c.client.AppsV1().Deployments(c.cfg.Namespace).Create(ctx, &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:   request.ComputeID,
			Labels: labels,
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
			Name:   request.ComputeID,
			Labels: labels,
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
		return fabric.RuntimeHandle{}, fmt.Errorf("create service: %w", err)
	}

	return fabric.RuntimeHandle{
		ProviderResourceID: "deployment/" + request.ComputeID,
		Status:             "running",
	}, nil
}

func (c *Client) CreateStorage(ctx context.Context, request fabric.CreateStorageRequest) (fabric.RuntimeHandle, error) {
	storageClass := c.cfg.StorageClass
	size := fmt.Sprintf("%dGi", request.Package.StorageGB)
	if request.Package.StorageGB <= 0 {
		size = "10Gi"
	}

	if _, err := c.client.CoreV1().PersistentVolumeClaims(c.cfg.Namespace).Create(ctx, &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:   request.StorageID,
			Labels: storageLabels(request.StorageID),
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes:      []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
			StorageClassName: &storageClass,
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
		ProviderResourceID: "pvc/" + request.StorageID,
		Status:             "available",
	}, nil
}

func (c *Client) AttachStorage(ctx context.Context, request fabric.AttachStorageRequest) (fabric.RuntimeHandle, error) {
	return fabric.RuntimeHandle{
		ProviderResourceID: request.AttachmentID,
		Status:             "attached",
	}, nil
}

func (c *Client) CreateWorkspaceRoute(ctx context.Context, request fabric.CreateRouteRequest) (fabric.RuntimeHandle, error) {
	if _, err := c.upsertTokenSecret(ctx, request.WorkspaceID, request.Token); err != nil {
		return fabric.RuntimeHandle{}, fmt.Errorf("upsert token secret: %w", err)
	}

	pathType := networkingv1.PathTypePrefix
	ingress := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:   routeName(request.WorkspaceID),
			Labels: routeLabels(request.WorkspaceID),
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
									Name: request.ComputeID,
									Port: networkingv1.ServiceBackendPort{Number: workspaceHTTPPort},
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
		return fabric.RuntimeHandle{}, fmt.Errorf("create ingress: %w", err)
	}

	return routeHandle(request.WorkspaceID, request.Token), nil
}

func (c *Client) DestroyCompute(ctx context.Context, request fabric.DestroyComputeRequest) error {
	deployments := c.client.AppsV1().Deployments(c.cfg.Namespace)
	services := c.client.CoreV1().Services(c.cfg.Namespace)

	return errors.Join(
		ignoreNotFound(deployments.Delete(ctx, request.ComputeID, metav1.DeleteOptions{})),
		ignoreNotFound(services.Delete(ctx, request.ComputeID, metav1.DeleteOptions{})),
	)
}

func (c *Client) DestroyStorage(ctx context.Context, request fabric.DestroyStorageRequest) error {
	return ignoreNotFound(c.client.CoreV1().PersistentVolumeClaims(c.cfg.Namespace).Delete(ctx, request.StorageID, metav1.DeleteOptions{}))
}

func (c *Client) DestroyWorkspaceRoute(ctx context.Context, request fabric.DestroyWorkspaceRouteRequest) error {
	return errors.Join(
		ignoreNotFound(c.client.NetworkingV1().Ingresses(c.cfg.Namespace).Delete(ctx, routeName(request.WorkspaceID), metav1.DeleteOptions{})),
		c.DeleteWorkspaceToken(ctx, fabric.DeleteWorkspaceTokenRequest{WorkspaceID: request.WorkspaceID}),
	)
}

func (c *Client) ResetWorkspaceToken(ctx context.Context, request fabric.ResetWorkspaceTokenRequest) (fabric.RuntimeHandle, error) {
	if _, err := c.upsertTokenSecret(ctx, request.WorkspaceID, request.Token); err != nil {
		return fabric.RuntimeHandle{}, fmt.Errorf("upsert token secret: %w", err)
	}
	return routeHandle(request.WorkspaceID, request.Token), nil
}

func (c *Client) DeleteWorkspaceToken(ctx context.Context, request fabric.DeleteWorkspaceTokenRequest) error {
	return ignoreNotFound(c.client.CoreV1().Secrets(c.cfg.Namespace).Delete(ctx, tokenSecretName(request.WorkspaceID), metav1.DeleteOptions{}))
}

func (c *Client) upsertTokenSecret(ctx context.Context, workspaceID, token string) (*corev1.Secret, error) {
	secrets := c.client.CoreV1().Secrets(c.cfg.Namespace)
	name := tokenSecretName(workspaceID)

	secret, err := secrets.Get(ctx, name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		return secrets.Create(ctx, tokenSecret(workspaceID, token), metav1.CreateOptions{})
	}
	if err != nil {
		return nil, err
	}

	secret.Data = map[string][]byte{"token": []byte(token)}
	if secret.Labels == nil {
		secret.Labels = routeLabels(workspaceID)
	}
	return secrets.Update(ctx, secret, metav1.UpdateOptions{})
}

func tokenSecret(workspaceID, token string) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:   tokenSecretName(workspaceID),
			Labels: routeLabels(workspaceID),
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
		"opl.one-person-lab/compute":  computeID,
	}
}

func storageLabels(storageID string) map[string]string {
	return map[string]string{
		"app.kubernetes.io/name":      "opl-workspace",
		"app.kubernetes.io/component": "storage",
		"opl.one-person-lab/storage":  storageID,
	}
}

func routeLabels(workspaceID string) map[string]string {
	return map[string]string{
		"app.kubernetes.io/name":       "opl-workspace",
		"app.kubernetes.io/component":  "route",
		"opl.one-person-lab/workspace": workspaceID,
	}
}

func tokenSecretName(workspaceID string) string {
	return "workspace-" + workspaceID + "-token"
}

func routeName(workspaceID string) string {
	return workspaceID
}

func workspacePath(workspaceID string) string {
	return "/w/" + workspaceID
}
