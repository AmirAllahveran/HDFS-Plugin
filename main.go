package main

import (
	"context"
	"flag"
	"fmt"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/client-go/util/homedir"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var rootCmd = &cobra.Command{
	Use:   "hdfs [cluster-name] [command]",
	Short: "Executes command on hadoop pod of given cluster.",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		var kubeconfig *string
		if home := homedir.HomeDir(); home != "" {
			kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeConfig file")
		} else {
			kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeConfig file")
		}
		flag.Parse()

		// Load the kubeConfig file as a Config
		config, err := clientcmd.LoadFromFile(*kubeconfig)
		if err != nil {
			panic(err)
		}

		// Build the rest.Config from the Config
		restConfig, err := clientcmd.NewDefaultClientConfig(*config, &clientcmd.ConfigOverrides{}).ClientConfig()
		if err != nil {
			panic(err)
		}

		// create the clientSet
		clientSet, err := kubernetes.NewForConfig(restConfig)
		if err != nil {
			panic(err)
		}

		namespace, _ := cmd.Flags().GetString("namespace")
		clusterName := args[0]
		command := strings.Split(args[1], " ")

		podList, err := clientSet.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		podExists := false
		for _, pod := range podList.Items {
			if strings.HasPrefix(pod.Name, clusterName+"-hadoop-") {
				execCommand(restConfig, clientSet, &pod, namespace, command)
				podExists = true
				break
			}
		}

		if !podExists {
			fmt.Println("No matching pod found.")
			os.Exit(1)
		}
	},
}

func main() {
	rootCmd.PersistentFlags().StringP("namespace", "n", "sample", "namespace")
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func execCommand(config *rest.Config, clientSet *kubernetes.Clientset, pod *v1.Pod, namespace string, command []string) {
	execRequest := clientSet.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(pod.Name).
		Namespace(namespace).
		SubResource("exec")

	execRequest.VersionedParams(&v1.PodExecOptions{
		Command:   command,
		Container: pod.Spec.Containers[0].Name, // Assuming command is to be executed in the first container
		Stdin:     false,
		Stdout:    true,
		Stderr:    true,
		TTY:       false,
	}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(config, "POST", execRequest.URL())
	if err != nil {
		fmt.Printf("Error while creating SPDY executor: %v\n", err)
		return
	}

	err = exec.StreamWithContext(context.Background(), remotecommand.StreamOptions{
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	})
	if err != nil {
		fmt.Printf("Error in Stream: %v\n", err)
		return
	}
}
