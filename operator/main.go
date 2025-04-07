package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	v1 "k8s.io/api/core/v1"
)

type EnvVar struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

func main() {
	// Load in-cluster configuration
	config, err := rest.InClusterConfig()
	if err != nil {
		fmt.Println("Failed to load in-cluster config:", err)
		return
	}

	// Create Kubernetes clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Println("Failed to create clientset:", err)
		return
	}

	fmt.Println("Operator started successfully. Watching ConfigMap 'environment'...")

	// Continuous sync loop
	for {
		err := syncEnvConfig(clientset)
		if err != nil {
			fmt.Println("Sync error:", err)
		}
		time.Sleep(10 * time.Second)
	}
}

func syncEnvConfig(clientset *kubernetes.Clientset) error {
	// to get the ConfigMap named "environment" from the default namespace
	cm, err := clientset.CoreV1().ConfigMaps("default").Get(context.TODO(), "environment", metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("could not read configmap: %v", err)
	}

	// Iterate over each key in the configmap
	for deployName, raw := range cm.Data {
		var envs []EnvVar
		if err := json.Unmarshal([]byte(raw), &envs); err != nil {
			fmt.Printf("Failed to parse JSON for '%s': %v\n", deployName, err)
			continue
		}

		// Get the Deployment matching the key
		deploy, err := clientset.AppsV1().Deployments("default").Get(context.TODO(), deployName, metav1.GetOptions{})
		if err != nil {
			fmt.Printf("Deployment '%s' not found: %v\n", deployName, err)
			continue
		}

		// Update env vars in first container
		container := &deploy.Spec.Template.Spec.Containers[0]
		for _, newEnv := range envs {
			found := false
			for i, existing := range container.Env {
				if existing.Name == newEnv.Name {
					container.Env[i].Value = newEnv.Value
					found = true
					break
				}
			}
			if !found {
				container.Env = append(container.Env, v1.EnvVar{Name: newEnv.Name, Value: newEnv.Value})
			}
		}

		// Apply the updated deployment
		_, err = clientset.AppsV1().Deployments("default").Update(context.TODO(), deploy, metav1.UpdateOptions{})
		if err != nil {
			fmt.Printf("Failed to update deployment '%s': %v\n", deployName, err)
		} else {
			fmt.Printf("Updated deployment: %s\n", deployName)
		}
	}
	return nil
}
