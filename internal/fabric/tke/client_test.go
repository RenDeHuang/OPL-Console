package tke

import (
	"context"
	"errors"
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/client-go/kubernetes/fake"
	clientgotesting "k8s.io/client-go/testing"

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

func TestCreateComputeRollsBackDeploymentWhenServiceCreateFails(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	clientset.PrependReactor("create", "services", func(action clientgotesting.Action) (bool, runtime.Object, error) {
		return true, nil, errors.New("service create failed")
	})
	client := New(testConfig(), clientset)

	_, err := client.CreateCompute(context.Background(), fabric.CreateComputeRequest{
		ComputeID:        "cmp-ws-alpha",
		BillingAccountID: "acct-owner",
		Package:          fabric.PackagePlan{ID: "basic", CPU: 2, MemoryGB: 4, StorageGB: 10},
	})
	if err == nil {
		t.Fatal("create compute error = nil")
	}
	if !strings.Contains(err.Error(), "create service") {
		t.Fatalf("error = %v, want wrapped service error", err)
	}

	deployments, err := client.client.AppsV1().Deployments("opl-cloud").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(deployments.Items) != 0 {
		t.Fatalf("deployments = %d, want rollback delete", len(deployments.Items))
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

func TestCreateStorageOmitsStorageClassWhenConfigIsEmpty(t *testing.T) {
	cfg := testConfig()
	cfg.StorageClass = ""
	client := New(cfg, fake.NewSimpleClientset())

	if _, err := client.CreateStorage(context.Background(), fabric.CreateStorageRequest{
		StorageID:        "stg-ws-alpha",
		BillingAccountID: "acct-owner",
		Package:          fabric.PackagePlan{ID: "basic", CPU: 2, MemoryGB: 4, StorageGB: 10},
	}); err != nil {
		t.Fatalf("create storage: %v", err)
	}

	pvc, err := client.client.CoreV1().PersistentVolumeClaims("opl-cloud").Get(context.Background(), "stg-ws-alpha", metav1.GetOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if pvc.Spec.StorageClassName != nil {
		t.Fatalf("storage class = %q, want nil", *pvc.Spec.StorageClassName)
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
	if got := ingress.Spec.Rules[0].HTTP.Paths[0].Backend.Service.Name; got != "opl-console" {
		t.Fatalf("backend service = %q, want console service", got)
	}
	if got := ingress.Spec.Rules[0].HTTP.Paths[0].Backend.Service.Port.Number; got != 8787 {
		t.Fatalf("backend service port = %d", got)
	}
	if got := ingress.Annotations["opl.one-person-lab/original-compute-id"]; got != "cmp-ws-alpha" {
		t.Fatalf("ingress original compute annotation = %q", got)
	}
}

func TestCreateWorkspaceRouteUsesConfiguredConsoleBackend(t *testing.T) {
	cfg := testConfig()
	cfg.ConsoleServiceName = "console-validator"
	cfg.ConsoleServicePort = 8080
	client := New(cfg, fake.NewSimpleClientset())

	if _, err := client.CreateWorkspaceRoute(context.Background(), fabric.CreateRouteRequest{
		WorkspaceID:   "ws-alpha",
		WorkspaceName: "Alpha",
		ComputeID:     "cmp-ws-alpha",
		Token:         "token-1",
	}); err != nil {
		t.Fatalf("create workspace route: %v", err)
	}

	ingress, err := client.client.NetworkingV1().Ingresses("opl-cloud").Get(context.Background(), "ws-alpha", metav1.GetOptions{})
	if err != nil {
		t.Fatalf("get ingress: %v", err)
	}
	backend := ingress.Spec.Rules[0].HTTP.Paths[0].Backend.Service
	if backend.Name != "console-validator" {
		t.Fatalf("backend service = %q", backend.Name)
	}
	if backend.Port.Number != 8080 {
		t.Fatalf("backend service port = %d", backend.Port.Number)
	}
}

func TestCreateWorkspaceRouteRollsBackTokenSecretWhenIngressCreateFails(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	clientset.PrependReactor("create", "ingresses", func(action clientgotesting.Action) (bool, runtime.Object, error) {
		return true, nil, errors.New("ingress create failed")
	})
	client := New(testConfig(), clientset)

	_, err := client.CreateWorkspaceRoute(context.Background(), fabric.CreateRouteRequest{
		WorkspaceID:   "ws-alpha",
		WorkspaceName: "Alpha",
		ComputeID:     "cmp-ws-alpha",
		Token:         "token-1",
	})
	if err == nil {
		t.Fatal("create workspace route error = nil")
	}
	if !strings.Contains(err.Error(), "create ingress") {
		t.Fatalf("error = %v, want wrapped ingress error", err)
	}

	if _, err := client.client.CoreV1().Secrets("opl-cloud").Get(context.Background(), "workspace-ws-alpha-token", metav1.GetOptions{}); !apierrors.IsNotFound(err) {
		t.Fatalf("secret get err = %v, want not found after rollback", err)
	}
}

func TestCreateWorkspaceRouteRestoresExistingTokenSecretWhenIngressCreateFails(t *testing.T) {
	clientset := fake.NewSimpleClientset(&corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "workspace-ws-alpha-token",
			Namespace: "opl-cloud",
			Labels: map[string]string{
				"existing-label": "keep",
			},
			Annotations: map[string]string{
				"existing-annotation": "keep",
			},
		},
		Data: map[string][]byte{"token": []byte("old-token")},
	})
	clientset.PrependReactor("create", "ingresses", func(action clientgotesting.Action) (bool, runtime.Object, error) {
		return true, nil, errors.New("ingress create failed")
	})
	client := New(testConfig(), clientset)

	_, err := client.CreateWorkspaceRoute(context.Background(), fabric.CreateRouteRequest{
		WorkspaceID:   "ws-alpha",
		WorkspaceName: "Alpha",
		ComputeID:     "cmp-ws-alpha",
		Token:         "new-token",
	})
	if err == nil {
		t.Fatal("create workspace route error = nil")
	}
	if !strings.Contains(err.Error(), "create ingress") {
		t.Fatalf("error = %v, want wrapped ingress error", err)
	}

	secret, err := client.client.CoreV1().Secrets("opl-cloud").Get(context.Background(), "workspace-ws-alpha-token", metav1.GetOptions{})
	if err != nil {
		t.Fatalf("get restored token secret: %v", err)
	}
	if got := string(secret.Data["token"]); got != "old-token" {
		t.Fatalf("secret token = %q", got)
	}
	if got := secret.Labels["existing-label"]; got != "keep" {
		t.Fatalf("existing label = %q", got)
	}
	if got := secret.Annotations["existing-annotation"]; got != "keep" {
		t.Fatalf("existing annotation = %q", got)
	}
}

func TestDNS1123NameAddsHashWhenSanitizationChangesValue(t *testing.T) {
	rawIDs := []string{"WS_ALPHA", "ws-alpha", "ws.alpha", "ws alpha"}
	names := map[string]string{}

	for _, rawID := range rawIDs {
		name := dns1123Name(rawID)
		assertDNS1123Name(t, name)
		if existingRawID, ok := names[name]; ok {
			t.Fatalf("dns1123Name(%q) and dns1123Name(%q) both produced %q", existingRawID, rawID, name)
		}
		names[name] = rawID
	}
}

func TestCreateMethodsUseDNS1123SafeResourceNamesAndAnnotateOriginalIDs(t *testing.T) {
	client := New(testConfig(), fake.NewSimpleClientset())
	ctx := context.Background()
	unsafeLongID := "WS_ALPHA.with_UNSAFE_chars_" + strings.Repeat("LongSegment", 8)
	computeID := "CMP_" + unsafeLongID
	storageID := "STG_" + unsafeLongID

	if _, err := client.CreateCompute(ctx, fabric.CreateComputeRequest{
		ComputeID:        computeID,
		BillingAccountID: "acct-owner",
		Package:          fabric.PackagePlan{ID: "basic", CPU: 2, MemoryGB: 4, StorageGB: 10},
	}); err != nil {
		t.Fatalf("create compute: %v", err)
	}
	if _, err := client.CreateStorage(ctx, fabric.CreateStorageRequest{
		StorageID:        storageID,
		BillingAccountID: "acct-owner",
		Package:          fabric.PackagePlan{ID: "basic", CPU: 2, MemoryGB: 4, StorageGB: 10},
	}); err != nil {
		t.Fatalf("create storage: %v", err)
	}
	if _, err := client.CreateWorkspaceRoute(ctx, fabric.CreateRouteRequest{
		WorkspaceID:   unsafeLongID,
		WorkspaceName: "Alpha",
		ComputeID:     computeID,
		Token:         "token-1",
	}); err != nil {
		t.Fatalf("create workspace route: %v", err)
	}

	deployments, err := client.client.AppsV1().Deployments("opl-cloud").List(ctx, metav1.ListOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(deployments.Items) != 1 {
		t.Fatalf("deployments = %d", len(deployments.Items))
	}
	assertDNS1123Name(t, deployments.Items[0].Name)
	if deployments.Items[0].Annotations["opl.one-person-lab/original-compute-id"] != computeID {
		t.Fatalf("deployment original compute annotation = %q", deployments.Items[0].Annotations["opl.one-person-lab/original-compute-id"])
	}

	services, err := client.client.CoreV1().Services("opl-cloud").List(ctx, metav1.ListOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(services.Items) != 1 {
		t.Fatalf("services = %d", len(services.Items))
	}
	assertDNS1123Name(t, services.Items[0].Name)
	if services.Items[0].Name != deployments.Items[0].Name {
		t.Fatalf("service name = %q, deployment name = %q", services.Items[0].Name, deployments.Items[0].Name)
	}

	pvcs, err := client.client.CoreV1().PersistentVolumeClaims("opl-cloud").List(ctx, metav1.ListOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(pvcs.Items) != 1 {
		t.Fatalf("pvcs = %d", len(pvcs.Items))
	}
	assertDNS1123Name(t, pvcs.Items[0].Name)
	if pvcs.Items[0].Annotations["opl.one-person-lab/original-storage-id"] != storageID {
		t.Fatalf("pvc original storage annotation = %q", pvcs.Items[0].Annotations["opl.one-person-lab/original-storage-id"])
	}

	ingresses, err := client.client.NetworkingV1().Ingresses("opl-cloud").List(ctx, metav1.ListOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(ingresses.Items) != 1 {
		t.Fatalf("ingresses = %d", len(ingresses.Items))
	}
	assertDNS1123Name(t, ingresses.Items[0].Name)
	if ingresses.Items[0].Annotations["opl.one-person-lab/original-workspace-id"] != unsafeLongID {
		t.Fatalf("ingress original workspace annotation = %q", ingresses.Items[0].Annotations["opl.one-person-lab/original-workspace-id"])
	}
	if got := ingresses.Items[0].Spec.Rules[0].HTTP.Paths[0].Backend.Service.Name; got != "opl-console" {
		t.Fatalf("backend service = %q, want console service", got)
	}
	if got := ingresses.Items[0].Annotations["opl.one-person-lab/original-compute-id"]; got != computeID {
		t.Fatalf("ingress original compute annotation = %q", got)
	}

	secrets, err := client.client.CoreV1().Secrets("opl-cloud").List(ctx, metav1.ListOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(secrets.Items) != 1 {
		t.Fatalf("secrets = %d", len(secrets.Items))
	}
	assertDNS1123Name(t, secrets.Items[0].Name)
	if secrets.Items[0].Annotations["opl.one-person-lab/original-workspace-id"] != unsafeLongID {
		t.Fatalf("secret original workspace annotation = %q", secrets.Items[0].Annotations["opl.one-person-lab/original-workspace-id"])
	}
	if secrets.Items[0].Annotations["opl.one-person-lab/original-compute-id"] != computeID {
		t.Fatalf("secret original compute annotation = %q", secrets.Items[0].Annotations["opl.one-person-lab/original-compute-id"])
	}
}

func TestDestroyComputeDeletesExistingDeploymentAndService(t *testing.T) {
	client := New(testConfig(), fake.NewSimpleClientset())
	ctx := context.Background()
	if _, err := client.CreateCompute(ctx, fabric.CreateComputeRequest{
		ComputeID:        "cmp-ws-alpha",
		BillingAccountID: "acct-owner",
		Package:          fabric.PackagePlan{ID: "basic", CPU: 2, MemoryGB: 4, StorageGB: 10},
	}); err != nil {
		t.Fatalf("create compute: %v", err)
	}

	if err := client.DestroyCompute(ctx, fabric.DestroyComputeRequest{ComputeID: "cmp-ws-alpha"}); err != nil {
		t.Fatalf("destroy compute: %v", err)
	}
	if _, err := client.client.AppsV1().Deployments("opl-cloud").Get(ctx, "cmp-ws-alpha", metav1.GetOptions{}); !apierrors.IsNotFound(err) {
		t.Fatalf("deployment get err = %v, want not found", err)
	}
	if _, err := client.client.CoreV1().Services("opl-cloud").Get(ctx, "cmp-ws-alpha", metav1.GetOptions{}); !apierrors.IsNotFound(err) {
		t.Fatalf("service get err = %v, want not found", err)
	}
}

func TestDestroyStorageDeletesExistingPVC(t *testing.T) {
	client := New(testConfig(), fake.NewSimpleClientset())
	ctx := context.Background()
	if _, err := client.CreateStorage(ctx, fabric.CreateStorageRequest{
		StorageID:        "stg-ws-alpha",
		BillingAccountID: "acct-owner",
		Package:          fabric.PackagePlan{ID: "basic", CPU: 2, MemoryGB: 4, StorageGB: 10},
	}); err != nil {
		t.Fatalf("create storage: %v", err)
	}

	if err := client.DestroyStorage(ctx, fabric.DestroyStorageRequest{StorageID: "stg-ws-alpha"}); err != nil {
		t.Fatalf("destroy storage: %v", err)
	}
	if _, err := client.client.CoreV1().PersistentVolumeClaims("opl-cloud").Get(ctx, "stg-ws-alpha", metav1.GetOptions{}); !apierrors.IsNotFound(err) {
		t.Fatalf("pvc get err = %v, want not found", err)
	}
}

func TestDestroyWorkspaceRouteDeletesExistingIngressAndTokenSecret(t *testing.T) {
	client := New(testConfig(), fake.NewSimpleClientset())
	ctx := context.Background()
	if _, err := client.CreateWorkspaceRoute(ctx, fabric.CreateRouteRequest{
		WorkspaceID:   "ws-alpha",
		WorkspaceName: "Alpha",
		ComputeID:     "cmp-ws-alpha",
		Token:         "token-1",
	}); err != nil {
		t.Fatalf("create workspace route: %v", err)
	}

	if err := client.DestroyWorkspaceRoute(ctx, fabric.DestroyWorkspaceRouteRequest{WorkspaceID: "ws-alpha"}); err != nil {
		t.Fatalf("destroy workspace route: %v", err)
	}
	if _, err := client.client.NetworkingV1().Ingresses("opl-cloud").Get(ctx, "ws-alpha", metav1.GetOptions{}); !apierrors.IsNotFound(err) {
		t.Fatalf("ingress get err = %v, want not found", err)
	}
	if _, err := client.client.CoreV1().Secrets("opl-cloud").Get(ctx, "workspace-ws-alpha-token", metav1.GetOptions{}); !apierrors.IsNotFound(err) {
		t.Fatalf("secret get err = %v, want not found", err)
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

func TestDestroyMethodsUseDNS1123SafeResourceNames(t *testing.T) {
	client := New(testConfig(), fake.NewSimpleClientset())
	ctx := context.Background()
	unsafeLongID := "WS_ALPHA.with_UNSAFE_chars_" + strings.Repeat("LongSegment", 8)
	computeID := "CMP_" + unsafeLongID
	storageID := "STG_" + unsafeLongID

	if _, err := client.CreateCompute(ctx, fabric.CreateComputeRequest{ComputeID: computeID, Package: fabric.PackagePlan{StorageGB: 10}}); err != nil {
		t.Fatalf("create compute: %v", err)
	}
	if _, err := client.CreateStorage(ctx, fabric.CreateStorageRequest{StorageID: storageID, Package: fabric.PackagePlan{StorageGB: 10}}); err != nil {
		t.Fatalf("create storage: %v", err)
	}
	if _, err := client.CreateWorkspaceRoute(ctx, fabric.CreateRouteRequest{WorkspaceID: unsafeLongID, ComputeID: computeID, Token: "token-1"}); err != nil {
		t.Fatalf("create workspace route: %v", err)
	}

	if err := client.DestroyCompute(ctx, fabric.DestroyComputeRequest{ComputeID: computeID}); err != nil {
		t.Fatalf("destroy compute: %v", err)
	}
	if err := client.DestroyStorage(ctx, fabric.DestroyStorageRequest{StorageID: storageID}); err != nil {
		t.Fatalf("destroy storage: %v", err)
	}
	if err := client.DestroyWorkspaceRoute(ctx, fabric.DestroyWorkspaceRouteRequest{WorkspaceID: unsafeLongID}); err != nil {
		t.Fatalf("destroy workspace route: %v", err)
	}

	assertNoRemainingResources(t, client)
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

func TestResetWorkspaceTokenMergesRequiredSecretLabelsAndAnnotations(t *testing.T) {
	clientset := fake.NewSimpleClientset(&corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      tokenSecretName("ws-alpha"),
			Namespace: "opl-cloud",
			Labels: map[string]string{
				"existing-label": "keep",
			},
			Annotations: map[string]string{
				"existing-annotation": "keep",
			},
		},
		Data: map[string][]byte{"token": []byte("old-token")},
	})
	client := New(testConfig(), clientset)

	if _, err := client.ResetWorkspaceToken(context.Background(), fabric.ResetWorkspaceTokenRequest{
		WorkspaceID: "ws-alpha",
		Token:       "token-2",
	}); err != nil {
		t.Fatalf("reset workspace token: %v", err)
	}

	secret, err := client.client.CoreV1().Secrets("opl-cloud").Get(context.Background(), tokenSecretName("ws-alpha"), metav1.GetOptions{})
	if err != nil {
		t.Fatalf("get token secret: %v", err)
	}
	if got := secret.Labels["existing-label"]; got != "keep" {
		t.Fatalf("existing label = %q", got)
	}
	if got := secret.Labels["app.kubernetes.io/component"]; got != "route" {
		t.Fatalf("required label = %q", got)
	}
	if got := secret.Annotations["existing-annotation"]; got != "keep" {
		t.Fatalf("existing annotation = %q", got)
	}
	if got := secret.Annotations["opl.one-person-lab/original-workspace-id"]; got != "ws-alpha" {
		t.Fatalf("required annotation = %q", got)
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

func assertDNS1123Name(t *testing.T, name string) {
	t.Helper()
	if errs := validation.IsDNS1123Label(name); len(errs) != 0 {
		t.Fatalf("name %q is not DNS-1123 safe: %v", name, errs)
	}
	if len(name) > 63 {
		t.Fatalf("name length = %d, want <= 63", len(name))
	}
}

func assertNoRemainingResources(t *testing.T, client *Client) {
	t.Helper()
	ctx := context.Background()

	deployments, err := client.client.AppsV1().Deployments("opl-cloud").List(ctx, metav1.ListOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(deployments.Items) != 0 {
		t.Fatalf("deployments = %d, want 0", len(deployments.Items))
	}

	services, err := client.client.CoreV1().Services("opl-cloud").List(ctx, metav1.ListOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(services.Items) != 0 {
		t.Fatalf("services = %d, want 0", len(services.Items))
	}

	pvcs, err := client.client.CoreV1().PersistentVolumeClaims("opl-cloud").List(ctx, metav1.ListOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(pvcs.Items) != 0 {
		t.Fatalf("pvcs = %d, want 0", len(pvcs.Items))
	}

	ingresses, err := client.client.NetworkingV1().Ingresses("opl-cloud").List(ctx, metav1.ListOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(ingresses.Items) != 0 {
		t.Fatalf("ingresses = %d, want 0", len(ingresses.Items))
	}

	secrets, err := client.client.CoreV1().Secrets("opl-cloud").List(ctx, metav1.ListOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(secrets.Items) != 0 {
		t.Fatalf("secrets = %d, want 0", len(secrets.Items))
	}
}
