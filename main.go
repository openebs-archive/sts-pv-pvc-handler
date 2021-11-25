package main

import (
	"context"
	"fmt"

	"github.com/ksraj123/lister-sa/pkg/constants"
	"github.com/ksraj123/lister-sa/pkg/executor"
	"github.com/ksraj123/lister-sa/pkg/utils"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var (
	clientset *kubernetes.Clientset
	ctx       context.Context
)

func init() {
	config, err := rest.InClusterConfig()
	if err != nil {
		fmt.Printf("error %s, getting inclusterconfig", err.Error())
	}
	clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Printf("error %s, creating clientset\n", err.Error())
	}
	ctx = context.Background()
}

func main() {
	namespaces := utils.EnvVarSlice(constants.NAMESPACES_ENV_VAR)
	for _, namespace := range namespaces {
		executor.Execute(clientset, ctx, namespace)
	}
}
