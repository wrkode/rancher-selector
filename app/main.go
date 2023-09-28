package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/retry"
)

type ProjectEvent struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
}

func main() {
	http.HandleFunc("/api/notify", handler)

	log.Println("Server starting on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST is supported", http.StatusMethodNotAllowed)
		return
	}

	var event ProjectEvent
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := triggerCronJob(os.Getenv("CRONJOB_NAME"), "kube-system"); err != nil {
		http.Error(w, "Failed to trigger cronjob", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("CronJob triggered successfully"))
}

func triggerCronJob(name, namespace string) error {
	config, err := rest.InClusterConfig()
	if err != nil {
		return err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		cronJob, err := clientset.BatchV1beta1().CronJobs(namespace).Get(context.TODO(), name, metav1.GetOptions{})
		if err != nil {
			return err
		}

		// Trigger the cronjob by modifying its annotations
		if cronJob.ObjectMeta.Annotations == nil {
			cronJob.ObjectMeta.Annotations = map[string]string{}
		}
		cronJob.ObjectMeta.Annotations["kubectl.kubernetes.io/restartedAt"] = time.Now().Format(time.RFC3339)

		_, updateErr := clientset.BatchV1beta1().CronJobs(namespace).Update(context.TODO(), cronJob, metav1.UpdateOptions{})
		return updateErr
	})
}
