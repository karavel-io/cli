package kubeclient

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/kustomize/api/resource"
)

// KubeClient is a simple abstraction over the dynamic Kubernetes Go client library
type KubeClient struct {
	config          *rest.Config
	client          *kubernetes.Clientset
	dynamicClient   dynamic.Interface
	discoveryClient *discovery.DiscoveryClient
	mapper          *restmapper.DeferredDiscoveryRESTMapper
	decoder         runtime.Serializer
	groups          map[string]metav1.APIGroup
}

// NewClientFromConfig initialized a client using a specified .kubeconfig
func NewClientFromConfig(configPath string) (*KubeClient, error) {
	config, err := clientcmd.BuildConfigFromFlags("", configPath)
	if err != nil {
		return nil, fmt.Errorf("could not load required information for kubernetes client: %w", err)
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("could not initialize kubernetes client: %w", err)
	}

	discoveryClient, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("could not initialize discovery client: %w", err)
	}

	return &KubeClient{
		config:          config,
		dynamicClient:   dynamicClient,
		discoveryClient: discoveryClient,
		mapper:          restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(discoveryClient)),
		decoder:         yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme),
		groups:          make(map[string]metav1.APIGroup),
	}, nil
}

func (k *KubeClient) FetchResources() error {
	groups, err := k.discoveryClient.ServerGroups()
	if err != nil {
		return err
	}
	for _, group := range groups.Groups {
		for _, version := range group.Versions {
			k.groups[version.GroupVersion] = group
		}
	}
	return nil
}

func (k *KubeClient) IsResourceSupported(resource *resource.Resource) bool {
	_, ok := k.groups[resource.GetApiVersion()]
	return ok
}

func (k *KubeClient) ApplyResource(ctx context.Context, resource *resource.Resource) error {
	yamlData, err := resource.AsYAML()
	if err != nil {
		return fmt.Errorf("could not encode to YAML: %w", err)
	}

	gvk := resource.GetGvk()
	namespace := resource.GetNamespace()

	mapping, err := k.mapper.RESTMapping(schema.GroupKind{Group: gvk.Group, Kind: gvk.Kind}, gvk.Version)
	if err != nil {
		return fmt.Errorf("could not derive GVR mapping for resource: %w", err)
	}

	var dr dynamic.ResourceInterface
	if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
		dr = k.dynamicClient.Resource(mapping.Resource).Namespace(namespace)
	} else {
		dr = k.dynamicClient.Resource(mapping.Resource)
	}

	_, err = dr.Patch(ctx, resource.GetName(), types.ApplyPatchType, yamlData, metav1.PatchOptions{
		FieldManager: "kubectl",
	})
	if err != nil {
		fmt.Println(mapping.Resource)
		return fmt.Errorf("patch in cluster failed: %w", err)
	}

	return nil
}
