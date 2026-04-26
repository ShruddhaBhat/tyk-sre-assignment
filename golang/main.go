package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// DeploymentHealth represents a deployment with mismatched desired and available replicas.
type DeploymentHealth struct {
	Namespace         string `json:"namespace"`
	Deployment        string `json:"deployment"`
	DesiredReplicas   int32  `json:"desiredReplicas"`
	AvailableReplicas int32  `json:"availableReplicas"`
}

type DeploymentHealthResponse struct {
	Healthy    bool               `json:"healthy"`
	Mismatches []DeploymentHealth `json:"mismatches"`
}

func main() {
	kubeconfig := flag.String("kubeconfig", "", "path to kubeconfig, leave empty for in-cluster")
	listenAddr := flag.String("address", ":8080", "HTTP server listen address")

	flag.Parse()

	kConfig, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}

	clientset, err := kubernetes.NewForConfig(kConfig)
	if err != nil {
		panic(err)
	}

	version, err := getKubernetesVersion(clientset)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Connected to Kubernetes %s\n", version)
	if err := startServer(*listenAddr, clientset); err != nil {
		panic(err)
	}

}

// getKubernetesVersion returns a string GitVersion of the Kubernetes server defined by the clientset.
//
// If it can't connect an error will be returned, which makes it useful to check connectivity.
func getKubernetesVersion(clientset kubernetes.Interface) (string, error) {
	version, err := clientset.Discovery().ServerVersion()
	if err != nil {
		return "", err
	}

	return version.String(), nil
}

// startServer launches an HTTP server with defined handlers and blocks until it's terminated or fails with an error.
//
// Expects a listenAddr to bind to.
func startServer(listenAddr string, clientset kubernetes.Interface) error {
	http.HandleFunc("/healthz", healthHandler)
	http.HandleFunc("/deploymentHealth", deploymentHealthHandler(clientset))
	http.HandleFunc("/apiServerConnectivity", apiServerConnectivityHandler(clientset))

	fmt.Printf("Server listening on %s\n", listenAddr)

	return http.ListenAndServe(listenAddr, nil)
}

// healthHandler responds with the health status of the application.
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)

	_, err := w.Write([]byte("ok"))
	if err != nil {
		fmt.Println("failed writing to response")
	}
}

// Check Deployment Replicas and compare with desired count, return an error if they don't match.
func deploymentHealthHandler(clientset kubernetes.Interface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mismatches, err := checkDeploymentHealth(clientset)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to check deployments: %v", err), http.StatusInternalServerError)
			return
		}

		response := DeploymentHealthResponse{
			Healthy:    len(mismatches) == 0,
			Mismatches: mismatches,
		}
		w.Header().Set("Content-Type", "application/json")

		// Return 200 if healthy, 207 Multi-Status if there are mismatches
		if !response.Healthy {
			w.WriteHeader(http.StatusMultiStatus)
		}

		if err := json.NewEncoder(w).Encode(response); err != nil {
			fmt.Println("failed to encode response")
		}
	}
}

func checkDeploymentHealth(clientset kubernetes.Interface) ([]DeploymentHealth, error) {
	cntx := context.Background()
	namespace := "" //all namespaces by default
	var mismatches []DeploymentHealth

	alldeployments, err := clientset.AppsV1().Deployments(namespace).List(cntx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("Failed to list deployments\n%v\n", err)
	}

	for _, deployment := range alldeployments.Items {
		desiredReplicas := *deployment.Spec.Replicas
		availableReplicas := deployment.Status.AvailableReplicas

		if desiredReplicas == availableReplicas {
			continue
		}
		mismatches = append(mismatches, DeploymentHealth{
			Namespace:         deployment.Namespace,
			Deployment:        deployment.Name,
			DesiredReplicas:   desiredReplicas,
			AvailableReplicas: availableReplicas,
		})
	}
	return mismatches, nil
}

func apiServerConnectivityHandler(clientset kubernetes.Interface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		connectionStatus, err := apiServerConnectivity(clientset)
		if err != nil {
			http.Error(w, fmt.Sprintf("API Server connectivity check failed: %v", err), http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, "API Server connectivity status: %s", connectionStatus)
	}
}

func apiServerConnectivity(clientset kubernetes.Interface) (string, error) {
	fmt.Println("Checking API Server connectivity")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	request := clientset.Discovery().RESTClient().Get()

	isApiServerAlive := request.AbsPath("/livez").Do(ctx)
	if isApiServerAlive.Error() != nil {
		return "Liveness check failed", isApiServerAlive.Error()
	}

	isApiServerReady := request.AbsPath("/readyz").Do(ctx)
	if isApiServerReady.Error() != nil {
		return "Readiness check failed", isApiServerReady.Error()
	}

	return "CONNECTED", nil
}
