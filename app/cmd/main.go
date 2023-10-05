package cmd

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"log"
	"regexp"
	"strings"
)

type ProjectEvent struct {
	Namespace   string            `json:"namespace"`
	Name        string            `json:"name"`
	ProjectID   string            `json:"projectId"`
	Annotations map[string]string `json:"annotations"`
}

func sanitizeKey(key string) string {
	// Replace invalid characters with a hyphen and remove leading invalid characters
	sanitized := regexp.MustCompile(`[^-a-zA-Z0-9_.]`).ReplaceAllString(key, "-")
	return strings.TrimLeft(sanitized, "-")
}

func CreateOrUpdateConfigMap(event ProjectEvent, cmName, namespace string) error {
	config, err := rest.InClusterConfig()
	if err != nil {
		return err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	cmClient := clientset.CoreV1().ConfigMaps(namespace)

	// Construct data map
	data := make(map[string]string)
	projectIdSanitized := sanitizeKey(event.Name) // Use event.Name instead of event.ProjectID
	annotationsKey := fmt.Sprintf("%s-annotations", projectIdSanitized)
	for k, v := range event.Annotations {
		annotationEntry := fmt.Sprintf("%s = %s", sanitizeKey(k), v)
		// Append each annotation to the data map under the projectId-annotations key
		if existingVal, ok := data[annotationsKey]; ok {
			data[annotationsKey] = existingVal + "\n" + annotationEntry
		} else {
			data[annotationsKey] = annotationEntry
		}
	}

	// Try to retrieve the existing ConfigMap
	cm, err := cmClient.Get(context.TODO(), cmName, metav1.GetOptions{})

	if err != nil {
		// If the ConfigMap doesn't exist, create a new one
		newCM := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name: cmName,
			},
			Data: data,
		}

		_, err = cmClient.Create(context.TODO(), newCM, metav1.CreateOptions{})
		if err != nil {
			return err
		}
	} else {
		// If the ConfigMap exists, update the entries related to this project
		cm.Data[annotationsKey] = data[annotationsKey]

		_, err = cmClient.Update(context.TODO(), cm, metav1.UpdateOptions{})
		if err != nil {
			return err
		}
	}

	return nil
}

func DeleteProjectFromConfigMap(event ProjectEvent) error {
	config, err := rest.InClusterConfig()
	if err != nil {
		return err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	cmClient := clientset.CoreV1().ConfigMaps("kube-system")

	cm, err := cmClient.Get(context.TODO(), "rancher-data", metav1.GetOptions{})
	if err != nil {
		return err
	}

	projectIdSanitized := sanitizeKey(event.Name)

	// Remove the project entry from the ConfigMap data
	annotationsKey := fmt.Sprintf("%s-annotations", projectIdSanitized)
	delete(cm.Data, annotationsKey)

	_, err = cmClient.Update(context.TODO(), cm, metav1.UpdateOptions{})
	if err != nil {
		log.Printf("Error updating ConfigMap after deleting project: %v", err)
		return err
	}

	log.Println("Successfully deleted project from ConfigMap:", event.Name)
	return nil
}
