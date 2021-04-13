package scaling

import (
	"bytes"
	"context"
	"encoding/json"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"log"
	"net/http"
	"os"
	"time"
)

var (
	user       string
	password   string
	gatewayUrl string
)

type Body struct {
	Replicas int `json:"replicas"`
}

func init() {

	var ok bool
	gatewayUrl, ok = os.LookupEnv("GATEWAY_URL")
	if !ok {
		log.Fatal("GATEWAY_URL environment variable not set")
	}

	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatal(err.Error())
	}

	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err.Error())
	}

	// retrieve the secret
	secrets := clientset.CoreV1().Secrets("openfaas")
	authSecret, err := secrets.Get(context.TODO(), "basic-auth", metav1.GetOptions{})
	if err != nil {
		log.Fatal(err.Error())
	}

	// set gateway authorization credentials
	data := authSecret.Data
	user = string(data["basic-auth-user"])
	password = string(data["basic-auth-password"])
}

func ScaleFunction(replicas int, functionName string) error {

	body, err := json.Marshal(Body{Replicas: replicas})
	if err != nil {
		log.Fatal(err)
	}

	client := &http.Client{
		Timeout: time.Second * 10,
	}

	url := gatewayUrl + "/system/scale-function/" + functionName
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	req.SetBasicAuth(user, password)
	response, err := client.Do(req)
	if err != nil {
		return err
	}

	err = response.Body.Close()
	if err != nil {
		return err
	}

	return nil
}
