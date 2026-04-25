package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	kubeconfig := flag.String("kubeconfig", "", "path to kubeconfig, leave empty for in-cluster")
	//	listenAddr := flag.String("address", ":8080", "HTTP server listen address")

	flag.Parse()

	kConfig, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}

	clientset, err := kubernetes.NewForConfig(kConfig)
	if err != nil {
		panic(err)
	}

	connectivity, err := apiServerConnectivity(clientset)
	if err != nil {
		fmt.Printf("API Server connectivity check failed\n")
		fmt.Printf("%s\n", connectivity)
		fmt.Printf("%v\n", err)
		panic(err)
	} else {
		fmt.Printf("API server is alive and ready!\n")
		fmt.Printf("Connection........%s\n", connectivity)

		version, err := getKubernetesVersion(clientset)
		if err != nil {
			panic(err)
		}
		fmt.Printf("Connected to Kubernetes %s\n", version)
	}

	if _, err := checkDeploymentReplicas(clientset); err != nil {
		fmt.Printf("%v\n", err)
	}

	//	if err := startServer(*listenAddr); err != nil {
	//		panic(err)
	//	}

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
//func startServer(listenAddr string) error {
//	http.HandleFunc("/healthz", healthHandler)

//	fmt.Printf("Server listening on %s\n", listenAddr)

//	return http.ListenAndServe(listenAddr, nil)
//}

// healthHandler responds with the health status of the application.
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)

	_, err := w.Write([]byte("ok"))
	if err != nil {
		fmt.Println("failed writing to response")
	}
}

// Check Deployment Replicas and compare with desired count, return an error if they don't match.
func checkDeploymentReplicas(clientset kubernetes.Interface) (string, error) {
	contxt := context.Background()
	namespace := "" //all namespaces by default

	alldeployments, err := clientset.AppsV1().Deployments(namespace).List(contxt, metav1.ListOptions{})
	if err != nil {
		return "", err
	}

	for _, deployment := range alldeployments.Items {
		if *deployment.Spec.Replicas != deployment.Status.ReadyReplicas {
			fmt.Printf("WARNING: Deployment replica check found mismatches \n %v", err)
			return "", fmt.Errorf("NAMESPACE:%s, NAME:%s, EXPECTED-REPLICAS:%d, READY-REPLICAS:%d", deployment.Namespace, deployment.Name, *deployment.Spec.Replicas, deployment.Status.ReadyReplicas)
		}
	}
	return "", nil
}

func apiServerConnectivity(clientset kubernetes.Interface) (string, error) {
	fmt.Println("Checking API Server connectivity")
	request := clientset.Discovery().RESTClient().Get()

	isApiServerAlive := request.AbsPath("/livez").Do(context.TODO())
	if isApiServerAlive.Error() != nil {
		return "Liveness check failed", isApiServerAlive.Error()
	}

	isApiServerReady := request.AbsPath("/readyz").Do(context.TODO())
	if isApiServerReady.Error() != nil {
		return "Readiness check failed", isApiServerReady.Error()
	}

	return "ESTABLISHED", nil
}
