package kube

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

// Metadata we want to extract for a K8s type
type ApiResource struct {
	Name string 
	Kind string 
	Namespaced bool 
	GVR schema.GroupVersionResource 
}

type DiscoveryClient struct {
	client discovery.DiscoveryInterface
}

func newDiscoveryClient(config *rest.Config) (*DiscoveryClient, error) {
	d, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return nil, err
	}

	return &DiscoveryClient{client: d}, nil
}

func (d *DiscoveryClient) getListableResources() ([]ApiResource, error) {
	lists, err := d.client.ServerPreferredResources()
	if err != nil {
		fmt.Printf("partial discovery results: %v\n", err)
	}

	var results []ApiResource
	for _, list := range lists {
		gv, err := schema.ParseGroupVersion(list.GroupVersion)
		if err != nil {
			continue
		}

		for _, res := range list.APIResources {
			if !slices.Contains(res.Verbs, "list") || strings.Contains(res.Name, "/") {
				continue
			}

			results = append(results, ApiResource{
				Name: res.Name,
				Kind: res.Kind,
				Namespaced: res.Namespaced,
				GVR: schema.GroupVersionResource{
					Group: gv.Group,
					Version: gv.Version,
					Resource: res.Name,
				},
			})
		}
	}

	return results, nil
}


func initializeK8sClientConfig() (*rest.Config, error) {
	kubeconfig := filepath.Join(homedir.HomeDir(), ".kube", "config")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func GetK8sDiscoveredResourcesList() ([]ApiResource, error) {
	config, err := initializeK8sClientConfig()
	if err != nil {
		return nil, err
	}

	discoveryClient, _ := newDiscoveryClient(config)
	resources, err := discoveryClient.getListableResources()
	if err != nil {
		return nil, err
	}

	return resources, nil
}
