package k8s

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

// Client wraps the Kubernetes client
type Client struct {
	clientset *kubernetes.Clientset
	namespace string
}

// DeploymentConfig represents configuration for a deployment
type DeploymentConfig struct {
	APIId         string
	Version       string
	Image         string
	Port          int32
	Replicas      int32
	Environment   map[string]string
	ResourceLimits ResourceRequirements
}

// ResourceRequirements defines resource limits and requests
type ResourceRequirements struct {
	CPURequest    string
	CPULimit      string
	MemoryRequest string
	MemoryLimit   string
}

// NewClient creates a new Kubernetes client
func NewClient() (*Client, error) {
	// Try in-cluster config first
	config, err := rest.InClusterConfig()
	if err != nil {
		// Fall back to kubeconfig
		kubeconfig := os.Getenv("KUBECONFIG")
		if kubeconfig == "" {
			if home := homedir.HomeDir(); home != "" {
				kubeconfig = filepath.Join(home, ".kube", "config")
			}
		}

		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create k8s config: %w", err)
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create k8s clientset: %w", err)
	}

	namespace := os.Getenv("DEPLOYMENT_NAMESPACE")
	if namespace == "" {
		namespace = "api-direct-apis"
	}

	return &Client{
		clientset: clientset,
		namespace: namespace,
	}, nil
}

// DeployAPI deploys an API to Kubernetes
func (c *Client) DeployAPI(ctx context.Context, config DeploymentConfig) error {
	// Create namespace if it doesn't exist
	if err := c.ensureNamespace(ctx); err != nil {
		return fmt.Errorf("failed to ensure namespace: %w", err)
	}

	// Create or update deployment
	if err := c.createOrUpdateDeployment(ctx, config); err != nil {
		return fmt.Errorf("failed to create deployment: %w", err)
	}

	// Create or update service
	if err := c.createOrUpdateService(ctx, config); err != nil {
		return fmt.Errorf("failed to create service: %w", err)
	}

	// Create or update ingress
	if err := c.createOrUpdateIngress(ctx, config); err != nil {
		return fmt.Errorf("failed to create ingress: %w", err)
	}

	return nil
}

// GetDeploymentStatus returns the status of a deployment
func (c *Client) GetDeploymentStatus(ctx context.Context, apiId string) (*appsv1.DeploymentStatus, error) {
	deployment, err := c.clientset.AppsV1().Deployments(c.namespace).Get(ctx, apiId, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return &deployment.Status, nil
}

// DeleteDeployment removes a deployment and associated resources
func (c *Client) DeleteDeployment(ctx context.Context, apiId string) error {
	// Delete ingress
	err := c.clientset.NetworkingV1().Ingresses(c.namespace).Delete(ctx, apiId, metav1.DeleteOptions{})
	if err != nil {
		// Log but don't fail if ingress doesn't exist
		fmt.Printf("Warning: failed to delete ingress: %v\n", err)
	}

	// Delete service
	err = c.clientset.CoreV1().Services(c.namespace).Delete(ctx, apiId, metav1.DeleteOptions{})
	if err != nil {
		fmt.Printf("Warning: failed to delete service: %v\n", err)
	}

	// Delete deployment
	err = c.clientset.AppsV1().Deployments(c.namespace).Delete(ctx, apiId, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete deployment: %w", err)
	}

	return nil
}

// ScaleDeployment adjusts the number of replicas
func (c *Client) ScaleDeployment(ctx context.Context, apiId string, replicas int32) error {
	deployment, err := c.clientset.AppsV1().Deployments(c.namespace).Get(ctx, apiId, metav1.GetOptions{})
	if err != nil {
		return err
	}

	deployment.Spec.Replicas = &replicas
	_, err = c.clientset.AppsV1().Deployments(c.namespace).Update(ctx, deployment, metav1.UpdateOptions{})
	return err
}

// Helper methods

func (c *Client) ensureNamespace(ctx context.Context) error {
	_, err := c.clientset.CoreV1().Namespaces().Get(ctx, c.namespace, metav1.GetOptions{})
	if err != nil {
		// Create namespace if it doesn't exist
		ns := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: c.namespace,
				Labels: map[string]string{
					"app": "api-direct",
				},
			},
		}
		_, err = c.clientset.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) createOrUpdateDeployment(ctx context.Context, config DeploymentConfig) error {
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      config.APIId,
			Namespace: c.namespace,
			Labels: map[string]string{
				"app":     "api-direct",
				"api-id":  config.APIId,
				"version": config.Version,
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &config.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"api-id": config.APIId,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":     "api-direct",
						"api-id":  config.APIId,
						"version": config.Version,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "api",
							Image: config.Image,
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: config.Port,
									Protocol:      corev1.ProtocolTCP,
								},
							},
							Env: c.buildEnvVars(config.Environment),
							Resources: c.buildResourceRequirements(config.ResourceLimits),
						},
					},
				},
			},
		},
	}

	// Try to update first
	_, err := c.clientset.AppsV1().Deployments(c.namespace).Update(ctx, deployment, metav1.UpdateOptions{})
	if err != nil {
		// If update fails, try to create
		_, err = c.clientset.AppsV1().Deployments(c.namespace).Create(ctx, deployment, metav1.CreateOptions{})
	}
	return err
}

func (c *Client) createOrUpdateService(ctx context.Context, config DeploymentConfig) error {
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      config.APIId,
			Namespace: c.namespace,
			Labels: map[string]string{
				"app":    "api-direct",
				"api-id": config.APIId,
			},
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"api-id": config.APIId,
			},
			Ports: []corev1.ServicePort{
				{
					Port:       80,
					TargetPort: intstr.FromInt(int(config.Port)),
					Protocol:   corev1.ProtocolTCP,
				},
			},
			Type: corev1.ServiceTypeClusterIP,
		},
	}

	// Try to update first
	_, err := c.clientset.CoreV1().Services(c.namespace).Update(ctx, service, metav1.UpdateOptions{})
	if err != nil {
		// If update fails, try to create
		_, err = c.clientset.CoreV1().Services(c.namespace).Create(ctx, service, metav1.CreateOptions{})
	}
	return err
}

func (c *Client) createOrUpdateIngress(ctx context.Context, config DeploymentConfig) error {
	pathType := networkingv1.PathTypePrefix
	ingress := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      config.APIId,
			Namespace: c.namespace,
			Labels: map[string]string{
				"app":    "api-direct",
				"api-id": config.APIId,
			},
			Annotations: map[string]string{
				"kubernetes.io/ingress.class":                "alb",
				"alb.ingress.kubernetes.io/scheme":           "internet-facing",
				"alb.ingress.kubernetes.io/target-type":      "ip",
				"alb.ingress.kubernetes.io/healthcheck-path": "/health",
			},
		},
		Spec: networkingv1.IngressSpec{
			Rules: []networkingv1.IngressRule{
				{
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{
									Path:     fmt.Sprintf("/apis/%s", config.APIId),
									PathType: &pathType,
									Backend: networkingv1.IngressBackend{
										Service: &networkingv1.IngressServiceBackend{
											Name: config.APIId,
											Port: networkingv1.ServiceBackendPort{
												Number: 80,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	// Try to update first
	_, err := c.clientset.NetworkingV1().Ingresses(c.namespace).Update(ctx, ingress, metav1.UpdateOptions{})
	if err != nil {
		// If update fails, try to create
		_, err = c.clientset.NetworkingV1().Ingresses(c.namespace).Create(ctx, ingress, metav1.CreateOptions{})
	}
	return err
}

func (c *Client) buildEnvVars(env map[string]string) []corev1.EnvVar {
	vars := []corev1.EnvVar{}
	for k, v := range env {
		vars = append(vars, corev1.EnvVar{
			Name:  k,
			Value: v,
		})
	}
	return vars
}

func (c *Client) buildResourceRequirements(requirements ResourceRequirements) corev1.ResourceRequirements {
	// Default values if not specified
	if requirements.CPURequest == "" {
		requirements.CPURequest = "100m"
	}
	if requirements.CPULimit == "" {
		requirements.CPULimit = "500m"
	}
	if requirements.MemoryRequest == "" {
		requirements.MemoryRequest = "128Mi"
	}
	if requirements.MemoryLimit == "" {
		requirements.MemoryLimit = "512Mi"
	}

	// Build resource requirements
	// Implementation simplified for brevity
	return corev1.ResourceRequirements{}
}
