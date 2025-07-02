package test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestKubernetesDeployments(t *testing.T) {
	t.Parallel()

	// Test backend deployment
	t.Run("validate backend deployment", func(t *testing.T) {
		namespace := fmt.Sprintf("test-%s", random.UniqueId())
		kubectlOptions := k8s.NewKubectlOptions("", "", namespace)

		// Create namespace
		k8s.CreateNamespace(t, kubectlOptions, namespace)
		defer k8s.DeleteNamespace(t, kubectlOptions, namespace)

		// Deploy backend
		k8s.KubectlApply(t, kubectlOptions, "../k8s/backend/deployment.yaml")
		defer k8s.KubectlDelete(t, kubectlOptions, "../k8s/backend/deployment.yaml")

		// Wait for deployment to be available
		k8s.WaitUntilDeploymentAvailable(t, kubectlOptions, "api-marketplace-backend", 10, 30*time.Second)

		// Validate deployment
		deployment := k8s.GetDeployment(t, kubectlOptions, "api-marketplace-backend")
		assert.Equal(t, int32(3), *deployment.Spec.Replicas)
		assert.Equal(t, "api-marketplace-backend", deployment.Name)

		// Validate pod template
		podTemplate := deployment.Spec.Template
		assert.Len(t, podTemplate.Spec.Containers, 1)
		container := podTemplate.Spec.Containers[0]
		assert.Equal(t, "backend", container.Name)
		assert.Equal(t, "api-marketplace-backend:latest", container.Image)

		// Validate resource limits
		assert.NotNil(t, container.Resources.Limits)
		assert.NotNil(t, container.Resources.Requests)

		// Validate health checks
		assert.NotNil(t, container.LivenessProbe)
		assert.Equal(t, "/health", container.LivenessProbe.HTTPGet.Path)
		assert.NotNil(t, container.ReadinessProbe)
		assert.Equal(t, "/ready", container.ReadinessProbe.HTTPGet.Path)

		// Validate environment variables
		envVars := make(map[string]string)
		for _, env := range container.Env {
			envVars[env.Name] = env.Value
		}
		assert.Contains(t, envVars, "NODE_ENV")
		assert.Equal(t, "production", envVars["NODE_ENV"])
	})

	// Test frontend deployment
	t.Run("validate frontend deployment", func(t *testing.T) {
		namespace := fmt.Sprintf("test-%s", random.UniqueId())
		kubectlOptions := k8s.NewKubectlOptions("", "", namespace)

		k8s.CreateNamespace(t, kubectlOptions, namespace)
		defer k8s.DeleteNamespace(t, kubectlOptions, namespace)

		k8s.KubectlApply(t, kubectlOptions, "../k8s/frontend/deployment.yaml")
		defer k8s.KubectlDelete(t, kubectlOptions, "../k8s/frontend/deployment.yaml")

		k8s.WaitUntilDeploymentAvailable(t, kubectlOptions, "api-marketplace-frontend", 10, 30*time.Second)

		deployment := k8s.GetDeployment(t, kubectlOptions, "api-marketplace-frontend")
		assert.Equal(t, int32(2), *deployment.Spec.Replicas)

		// Validate anti-affinity rules
		affinity := deployment.Spec.Template.Spec.Affinity
		assert.NotNil(t, affinity)
		assert.NotNil(t, affinity.PodAntiAffinity)
		assert.Len(t, affinity.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution, 1)
	})
}

func TestKubernetesServices(t *testing.T) {
	t.Parallel()

	// Test service configurations
	t.Run("validate service definitions", func(t *testing.T) {
		namespace := fmt.Sprintf("test-%s", random.UniqueId())
		kubectlOptions := k8s.NewKubectlOptions("", "", namespace)

		k8s.CreateNamespace(t, kubectlOptions, namespace)
		defer k8s.DeleteNamespace(t, kubectlOptions, namespace)

		// Apply services
		k8s.KubectlApply(t, kubectlOptions, "../k8s/backend/service.yaml")
		defer k8s.KubectlDelete(t, kubectlOptions, "../k8s/backend/service.yaml")

		// Get and validate backend service
		service := k8s.GetService(t, kubectlOptions, "api-marketplace-backend")
		assert.Equal(t, corev1.ServiceTypeClusterIP, service.Spec.Type)
		assert.Len(t, service.Spec.Ports, 1)
		assert.Equal(t, int32(80), service.Spec.Ports[0].Port)
		assert.Equal(t, int32(3000), service.Spec.Ports[0].TargetPort.IntVal)

		// Validate selector
		assert.Equal(t, "api-marketplace-backend", service.Spec.Selector["app"])
	})

	// Test headless service
	t.Run("validate headless service for stateful components", func(t *testing.T) {
		namespace := fmt.Sprintf("test-%s", random.UniqueId())
		kubectlOptions := k8s.NewKubectlOptions("", "", namespace)

		k8s.CreateNamespace(t, kubectlOptions, namespace)
		defer k8s.DeleteNamespace(t, kubectlOptions, namespace)

		k8s.KubectlApply(t, kubectlOptions, "../k8s/database/service.yaml")
		defer k8s.KubectlDelete(t, kubectlOptions, "../k8s/database/service.yaml")

		service := k8s.GetService(t, kubectlOptions, "postgres-headless")
		assert.Equal(t, "None", service.Spec.ClusterIP)
		assert.Equal(t, corev1.ServiceTypeClusterIP, service.Spec.Type)
	})
}

func TestKubernetesIngress(t *testing.T) {
	t.Parallel()

	// Test ingress configuration
	t.Run("validate ingress rules", func(t *testing.T) {
		namespace := fmt.Sprintf("test-%s", random.UniqueId())
		kubectlOptions := k8s.NewKubectlOptions("", "", namespace)

		k8s.CreateNamespace(t, kubectlOptions, namespace)
		defer k8s.DeleteNamespace(t, kubectlOptions, namespace)

		// Apply ingress
		k8s.KubectlApply(t, kubectlOptions, "../k8s/ingress/ingress.yaml")
		defer k8s.KubectlDelete(t, kubectlOptions, "../k8s/ingress/ingress.yaml")

		// Get ingress
		ingress := k8s.GetIngress(t, kubectlOptions, "api-marketplace-ingress")

		// Validate ingress class
		assert.Equal(t, "nginx", *ingress.Spec.IngressClassName)

		// Validate TLS configuration
		assert.Len(t, ingress.Spec.TLS, 1)
		assert.Contains(t, ingress.Spec.TLS[0].Hosts, "api.apidirect.dev")
		assert.Equal(t, "api-marketplace-tls", ingress.Spec.TLS[0].SecretName)

		// Validate rules
		assert.Len(t, ingress.Spec.Rules, 2) // api and www subdomains
		
		// Validate API rule
		apiRule := ingress.Spec.Rules[0]
		assert.Equal(t, "api.apidirect.dev", apiRule.Host)
		assert.Len(t, apiRule.HTTP.Paths, 1)
		assert.Equal(t, "/", apiRule.HTTP.Paths[0].Path)
		assert.Equal(t, "api-marketplace-backend", apiRule.HTTP.Paths[0].Backend.Service.Name)

		// Validate annotations
		annotations := ingress.Annotations
		assert.Equal(t, "letsencrypt-prod", annotations["cert-manager.io/cluster-issuer"])
		assert.Equal(t, "true", annotations["nginx.ingress.kubernetes.io/ssl-redirect"])
		assert.Equal(t, "true", annotations["nginx.ingress.kubernetes.io/force-ssl-redirect"])
	})
}

func TestKubernetesConfigMaps(t *testing.T) {
	t.Parallel()

	// Test ConfigMap creation and mounting
	t.Run("validate ConfigMap usage", func(t *testing.T) {
		namespace := fmt.Sprintf("test-%s", random.UniqueId())
		kubectlOptions := k8s.NewKubectlOptions("", "", namespace)

		k8s.CreateNamespace(t, kubectlOptions, namespace)
		defer k8s.DeleteNamespace(t, kubectlOptions, namespace)

		// Apply ConfigMap
		k8s.KubectlApply(t, kubectlOptions, "../k8s/backend/configmap.yaml")
		defer k8s.KubectlDelete(t, kubectlOptions, "../k8s/backend/configmap.yaml")

		// Get ConfigMap
		configMap := k8s.GetConfigMap(t, kubectlOptions, "api-marketplace-config")
		
		// Validate data
		assert.Contains(t, configMap.Data, "app.properties")
		assert.Contains(t, configMap.Data, "redis.conf")
		
		// Validate immutability
		assert.True(t, *configMap.Immutable)
	})
}

func TestKubernetesSecrets(t *testing.T) {
	t.Parallel()

	// Test Secret management
	t.Run("validate Secret handling", func(t *testing.T) {
		namespace := fmt.Sprintf("test-%s", random.UniqueId())
		kubectlOptions := k8s.NewKubectlOptions("", "", namespace)

		k8s.CreateNamespace(t, kubectlOptions, namespace)
		defer k8s.DeleteNamespace(t, kubectlOptions, namespace)

		// Create a test secret
		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "api-marketplace-secrets",
				Namespace: namespace,
			},
			Type: corev1.SecretTypeOpaque,
			StringData: map[string]string{
				"database-url": "postgresql://user:pass@postgres:5432/db",
				"jwt-secret":   "test-jwt-secret",
				"api-key":      "test-api-key",
			},
		}

		k8s.KubectlApply(t, kubectlOptions, "../k8s/backend/secrets.yaml")
		defer k8s.KubectlDelete(t, kubectlOptions, "../k8s/backend/secrets.yaml")

		// Verify secret exists and is properly typed
		retrievedSecret := k8s.GetSecret(t, kubectlOptions, "api-marketplace-secrets")
		assert.Equal(t, corev1.SecretTypeOpaque, retrievedSecret.Type)
		assert.Contains(t, retrievedSecret.Data, "database-url")
		assert.Contains(t, retrievedSecret.Data, "jwt-secret")
	})
}

func TestKubernetesStatefulSets(t *testing.T) {
	t.Parallel()

	// Test StatefulSet for database
	t.Run("validate PostgreSQL StatefulSet", func(t *testing.T) {
		namespace := fmt.Sprintf("test-%s", random.UniqueId())
		kubectlOptions := k8s.NewKubectlOptions("", "", namespace)

		k8s.CreateNamespace(t, kubectlOptions, namespace)
		defer k8s.DeleteNamespace(t, kubectlOptions, namespace)

		// Apply StatefulSet
		k8s.KubectlApply(t, kubectlOptions, "../k8s/database/statefulset.yaml")
		defer k8s.KubectlDelete(t, kubectlOptions, "../k8s/database/statefulset.yaml")

		// Wait for StatefulSet to be ready
		k8s.WaitUntilStatefulSetAvailable(t, kubectlOptions, "postgres", 5, 30*time.Second)

		// Get StatefulSet
		statefulSet := getStatefulSet(t, kubectlOptions, "postgres")
		
		// Validate configuration
		assert.Equal(t, int32(1), *statefulSet.Spec.Replicas)
		assert.Equal(t, "postgres", statefulSet.Spec.ServiceName)
		
		// Validate volume claim templates
		assert.Len(t, statefulSet.Spec.VolumeClaimTemplates, 1)
		volumeClaim := statefulSet.Spec.VolumeClaimTemplates[0]
		assert.Equal(t, "postgres-data", volumeClaim.Name)
		assert.Equal(t, "10Gi", volumeClaim.Spec.Resources.Requests.Storage().String())
		
		// Validate pod management policy
		assert.Equal(t, appsv1.OrderedReadyPodManagement, statefulSet.Spec.PodManagementPolicy)
	})
}

func TestKubernetesHorizontalPodAutoscaler(t *testing.T) {
	t.Parallel()

	// Test HPA configuration
	t.Run("validate HPA for backend", func(t *testing.T) {
		namespace := fmt.Sprintf("test-%s", random.UniqueId())
		kubectlOptions := k8s.NewKubectlOptions("", "", namespace)

		k8s.CreateNamespace(t, kubectlOptions, namespace)
		defer k8s.DeleteNamespace(t, kubectlOptions, namespace)

		// Apply HPA
		k8s.KubectlApply(t, kubectlOptions, "../k8s/backend/hpa.yaml")
		defer k8s.KubectlDelete(t, kubectlOptions, "../k8s/backend/hpa.yaml")

		// Validate HPA configuration
		output := k8s.RunKubectl(t, kubectlOptions, "get", "hpa", "api-marketplace-backend", "-o", "json")
		
		// Parse and validate HPA settings
		assert.Contains(t, output, "\"minReplicas\": 2")
		assert.Contains(t, output, "\"maxReplicas\": 10")
		assert.Contains(t, output, "\"targetCPUUtilizationPercentage\": 70")
	})
}

func TestKubernetesPodDisruptionBudget(t *testing.T) {
	t.Parallel()

	// Test PDB configuration
	t.Run("validate PodDisruptionBudget", func(t *testing.T) {
		namespace := fmt.Sprintf("test-%s", random.UniqueId())
		kubectlOptions := k8s.NewKubectlOptions("", "", namespace)

		k8s.CreateNamespace(t, kubectlOptions, namespace)
		defer k8s.DeleteNamespace(t, kubectlOptions, namespace)

		// Apply PDB
		k8s.KubectlApply(t, kubectlOptions, "../k8s/backend/pdb.yaml")
		defer k8s.KubectlDelete(t, kubectlOptions, "../k8s/backend/pdb.yaml")

		// Validate PDB
		output := k8s.RunKubectl(t, kubectlOptions, "get", "pdb", "api-marketplace-backend", "-o", "json")
		
		// Should maintain at least 1 replica during disruptions
		assert.Contains(t, output, "\"minAvailable\": 1")
		assert.Contains(t, output, "\"app\": \"api-marketplace-backend\"")
	})
}

func TestKubernetesNetworkPolicies(t *testing.T) {
	t.Parallel()

	// Test network policies
	t.Run("validate network isolation", func(t *testing.T) {
		namespace := fmt.Sprintf("test-%s", random.UniqueId())
		kubectlOptions := k8s.NewKubectlOptions("", "", namespace)

		k8s.CreateNamespace(t, kubectlOptions, namespace)
		defer k8s.DeleteNamespace(t, kubectlOptions, namespace)

		// Apply network policy
		k8s.KubectlApply(t, kubectlOptions, "../k8s/network/network-policy.yaml")
		defer k8s.KubectlDelete(t, kubectlOptions, "../k8s/network/network-policy.yaml")

		// Get network policy
		networkPolicy := getNetworkPolicy(t, kubectlOptions, "api-marketplace-network-policy")
		
		// Validate ingress rules
		assert.Len(t, networkPolicy.Spec.Ingress, 2)
		
		// Frontend can access backend
		assert.Contains(t, networkPolicy.Spec.Ingress[0].From[0].PodSelector.MatchLabels, "app")
		assert.Equal(t, "api-marketplace-frontend", networkPolicy.Spec.Ingress[0].From[0].PodSelector.MatchLabels["app"])
		
		// Validate egress rules
		assert.Len(t, networkPolicy.Spec.Egress, 3) // DNS, Database, External APIs
	})
}

func TestKubernetesRBAC(t *testing.T) {
	t.Parallel()

	// Test RBAC configuration
	t.Run("validate service account and roles", func(t *testing.T) {
		namespace := fmt.Sprintf("test-%s", random.UniqueId())
		kubectlOptions := k8s.NewKubectlOptions("", "", namespace)

		k8s.CreateNamespace(t, kubectlOptions, namespace)
		defer k8s.DeleteNamespace(t, kubectlOptions, namespace)

		// Apply RBAC resources
		k8s.KubectlApply(t, kubectlOptions, "../k8s/rbac/service-account.yaml")
		defer k8s.KubectlDelete(t, kubectlOptions, "../k8s/rbac/service-account.yaml")

		k8s.KubectlApply(t, kubectlOptions, "../k8s/rbac/role.yaml")
		defer k8s.KubectlDelete(t, kubectlOptions, "../k8s/rbac/role.yaml")

		k8s.KubectlApply(t, kubectlOptions, "../k8s/rbac/role-binding.yaml")
		defer k8s.KubectlDelete(t, kubectlOptions, "../k8s/rbac/role-binding.yaml")

		// Validate service account exists
		serviceAccount := k8s.GetServiceAccount(t, kubectlOptions, "api-marketplace-backend")
		assert.NotNil(t, serviceAccount)

		// Validate role permissions
		output := k8s.RunKubectl(t, kubectlOptions, "get", "role", "api-marketplace-role", "-o", "json")
		
		// Should only have necessary permissions
		assert.Contains(t, output, "\"get\"")
		assert.Contains(t, output, "\"list\"")
		assert.Contains(t, output, "\"watch\"")
		assert.NotContains(t, output, "\"*\"") // No wildcard permissions
	})
}

func TestKubernetesResourceQuotas(t *testing.T) {
	t.Parallel()

	// Test resource quotas
	t.Run("validate namespace resource limits", func(t *testing.T) {
		namespace := fmt.Sprintf("test-%s", random.UniqueId())
		kubectlOptions := k8s.NewKubectlOptions("", "", namespace)

		k8s.CreateNamespace(t, kubectlOptions, namespace)
		defer k8s.DeleteNamespace(t, kubectlOptions, namespace)

		// Apply resource quota
		k8s.KubectlApply(t, kubectlOptions, "../k8s/quotas/resource-quota.yaml")
		defer k8s.KubectlDelete(t, kubectlOptions, "../k8s/quotas/resource-quota.yaml")

		// Get resource quota
		output := k8s.RunKubectl(t, kubectlOptions, "get", "resourcequota", "api-marketplace-quota", "-o", "json")
		
		// Validate limits
		assert.Contains(t, output, "\"requests.cpu\": \"10\"")
		assert.Contains(t, output, "\"requests.memory\": \"20Gi\"")
		assert.Contains(t, output, "\"limits.cpu\": \"20\"")
		assert.Contains(t, output, "\"limits.memory\": \"40Gi\"")
		assert.Contains(t, output, "\"persistentvolumeclaims\": \"10\"")
	})
}

// Helper function to get StatefulSet
func getStatefulSet(t *testing.T, options *k8s.KubectlOptions, name string) *appsv1.StatefulSet {
	clientset, err := k8s.GetKubernetesClientFromOptionsE(t, options)
	require.NoError(t, err)

	statefulSet, err := clientset.AppsV1().StatefulSets(options.Namespace).Get(context.Background(), name, metav1.GetOptions{})
	require.NoError(t, err)

	return statefulSet
}

// Helper function to get NetworkPolicy
func getNetworkPolicy(t *testing.T, options *k8s.KubectlOptions, name string) *networkingv1.NetworkPolicy {
	clientset, err := k8s.GetKubernetesClientFromOptionsE(t, options)
	require.NoError(t, err)

	networkPolicy, err := clientset.NetworkingV1().NetworkPolicies(options.Namespace).Get(context.Background(), name, metav1.GetOptions{})
	require.NoError(t, err)

	return networkPolicy
}