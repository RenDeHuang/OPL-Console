package tke

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/RenDeHuang/opl-console/internal/fabric"
)

func TestCreateComputeCreatesDeploymentAndService(t *testing.T) {
	client := New(testConfig(), fake.NewSimpleClientset())

	handle, err := client.CreateCompute(context.Background(), fabric.CreateComputeRequest{
		ComputeID:        "cmp-ws-alpha",
		BillingAccountID: "acct-owner",
		Package:          fabric.PackagePlan{ID: "basic", CPU: 2, MemoryGB: 4, StorageGB: 10},
	})
	if err != nil {
		t.Fatalf("create compute: %v", err)
	}
	if handle.ProviderResourceID != "deployment/cmp-ws-alpha" {
		t.Fatalf("provider resource id = %q", handle.ProviderResourceID)
	}

	deployments, err := client.client.AppsV1().Deployments("opl-cloud").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(deployments.Items) != 1 {
		t.Fatalf("deployments = %d", len(deployments.Items))
	}
	deployment := deployments.Items[0]
	if deployment.Name != "cmp-ws-alpha" {
		t.Fatalf("deployment name = %q", deployment.Name)
	}
	if deployment.Spec.Template.Spec.Containers[0].Image != testConfig().Image {
		t.Fatalf("deployment image = %q", deployment.Spec.Template.Spec.Containers[0].Image)
	}

	services, err := client.client.CoreV1().Services("opl-cloud").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(services.Items) != 1 {
		t.Fatalf("services = %d", len(services.Items))
	}
	service := services.Items[0]
	if service.Name != "cmp-ws-alpha" {
		t.Fatalf("service name = %q", service.Name)
	}
	if service.Spec.Type != corev1.ServiceTypeClusterIP {
		t.Fatalf("service type = %s", service.Spec.Type)
	}
}

func TestCreateStorageCreatesPVC(t *testing.T) {
	client := New(testConfig(), fake.NewSimpleClientset())

	handle, err := client.CreateStorage(context.Background(), fabric.CreateStorageRequest{
		StorageID:        "stg-ws-alpha",
		BillingAccountID: "acct-owner",
		Package:          fabric.PackagePlan{ID: "basic", CPU: 2, MemoryGB: 4, StorageGB: 10},
	})
	if err != nil {
		t.Fatalf("create storage: %v", err)
	}
	if handle.ProviderResourceID != "pvc/stg-ws-alpha" {
		t.Fatalf("provider resource id = %q", handle.ProviderResourceID)
	}

	pvcs, err := client.client.CoreV1().PersistentVolumeClaims("opl-cloud").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(pvcs.Items) != 1 {
		t.Fatalf("pvcs = %d", len(pvcs.Items))
	}
	pvc := pvcs.Items[0]
	if pvc.Name != "stg-ws-alpha" {
		t.Fatalf("pvc name = %q", pvc.Name)
	}
	if pvc.Spec.StorageClassName == nil || *pvc.Spec.StorageClassName != "cbs" {
		t.Fatalf("storage class = %v", pvc.Spec.StorageClassName)
	}
	if got := pvc.Spec.Resources.Requests.Storage().String(); got != "10Gi" {
		t.Fatalf("storage request = %q", got)
	}
}

func TestCreateWorkspaceRouteCreatesTokenSecretAndIngress(t *testing.T) {
	client := New(testConfig(), fake.NewSimpleClientset())

	handle, err := client.CreateWorkspaceRoute(context.Background(), fabric.CreateRouteRequest{
		WorkspaceID:   "ws-alpha",
		WorkspaceName: "Alpha",
		ComputeID:     "cmp-ws-alpha",
		Token:         "token-1",
	})
	if err != nil {
		t.Fatalf("create workspace route: %v", err)
	}
	if handle.ProviderResourceID != "ingress/ws-alpha" {
		t.Fatalf("provider resource id = %q", handle.ProviderResourceID)
	}
	if handle.URL != "/w/ws-alpha?token=token-1" {
		t.Fatalf("url = %q", handle.URL)
	}

	secret, err := client.client.CoreV1().Secrets("opl-cloud").Get(context.Background(), "workspace-ws-alpha-token", metav1.GetOptions{})
	if err != nil {
		t.Fatalf("get token secret: %v", err)
	}
	if got := string(secret.Data["token"]); got != "token-1" {
		t.Fatalf("secret token = %q", got)
	}

	ingress, err := client.client.NetworkingV1().Ingresses("opl-cloud").Get(context.Background(), "ws-alpha", metav1.GetOptions{})
	if err != nil {
		t.Fatalf("get ingress: %v", err)
	}
	if ingress.Spec.IngressClassName == nil || *ingress.Spec.IngressClassName != "nginx" {
		t.Fatalf("ingress class = %v", ingress.Spec.IngressClassName)
	}
	if got := ingress.Spec.Rules[0].HTTP.Paths[0].Path; got != "/w/ws-alpha" {
		t.Fatalf("ingress path = %q", got)
	}
	if got := ingress.Spec.Rules[0].HTTP.Paths[0].Backend.Service.Name; got != "cmp-ws-alpha" {
		t.Fatalf("backend service = %q", got)
	}
}

func TestDestroyMethodsTreatMissingResourcesAsSuccess(t *testing.T) {
	client := New(testConfig(), fake.NewSimpleClientset())
	ctx := context.Background()

	if err := client.DestroyCompute(ctx, fabric.DestroyComputeRequest{ComputeID: "missing-compute"}); err != nil {
		t.Fatalf("destroy missing compute: %v", err)
	}
	if err := client.DestroyStorage(ctx, fabric.DestroyStorageRequest{StorageID: "missing-storage"}); err != nil {
		t.Fatalf("destroy missing storage: %v", err)
	}
	if err := client.DestroyWorkspaceRoute(ctx, fabric.DestroyWorkspaceRouteRequest{WorkspaceID: "missing-workspace"}); err != nil {
		t.Fatalf("destroy missing route: %v", err)
	}
}

func TestResetWorkspaceTokenCreatesUpdatesSecretAndReturnsURL(t *testing.T) {
	client := New(testConfig(), fake.NewSimpleClientset())
	ctx := context.Background()

	handle, err := client.ResetWorkspaceToken(ctx, fabric.ResetWorkspaceTokenRequest{
		WorkspaceID: "ws-alpha",
		Token:       "token-1",
	})
	if err != nil {
		t.Fatalf("reset workspace token: %v", err)
	}
	if handle.URL != "/w/ws-alpha?token=token-1" {
		t.Fatalf("url = %q", handle.URL)
	}

	handle, err = client.ResetWorkspaceToken(ctx, fabric.ResetWorkspaceTokenRequest{
		WorkspaceID: "ws-alpha",
		Token:       "token-2",
	})
	if err != nil {
		t.Fatalf("reset workspace token again: %v", err)
	}
	if handle.URL != "/w/ws-alpha?token=token-2" {
		t.Fatalf("url = %q", handle.URL)
	}

	secret, err := client.client.CoreV1().Secrets("opl-cloud").Get(ctx, "workspace-ws-alpha-token", metav1.GetOptions{})
	if err != nil {
		t.Fatalf("get token secret: %v", err)
	}
	if got := string(secret.Data["token"]); got != "token-2" {
		t.Fatalf("secret token = %q", got)
	}
}

func TestDeleteWorkspaceTokenIsIdempotent(t *testing.T) {
	client := New(testConfig(), fake.NewSimpleClientset())
	ctx := context.Background()

	if _, err := client.ResetWorkspaceToken(ctx, fabric.ResetWorkspaceTokenRequest{WorkspaceID: "ws-alpha", Token: "token-1"}); err != nil {
		t.Fatalf("reset workspace token: %v", err)
	}
	if err := client.DeleteWorkspaceToken(ctx, fabric.DeleteWorkspaceTokenRequest{WorkspaceID: "ws-alpha"}); err != nil {
		t.Fatalf("delete workspace token: %v", err)
	}
	if err := client.DeleteWorkspaceToken(ctx, fabric.DeleteWorkspaceTokenRequest{WorkspaceID: "ws-alpha"}); err != nil {
		t.Fatalf("delete missing workspace token: %v", err)
	}
}

func testConfig() Config {
	return Config{
		Namespace:    "opl-cloud",
		Image:        "ghcr.io/gaofeng21cn/one-person-lab-app:latest",
		StorageClass: "cbs",
		IngressClass: "nginx",
	}
}
