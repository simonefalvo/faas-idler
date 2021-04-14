package scaling

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
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

type functionSpec struct {
	Name     string            `json:"name"`
	Replicas int               `json:"replicas"`
	Labels   map[string]string `json:"labels,omitempty"`
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

func ScalableFunctions() ([]string, error) {

	// make http api request
	url := gatewayUrl + "/system/functions"
	resBody, err := apiRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	var functions []functionSpec
	err = json.Unmarshal(resBody, &functions)
	if err != nil {
		return nil, err
	}

	var scalableFunctions []string
	for _, f := range functions {
		labels := f.Labels
		replicas := f.Replicas
		if labels != nil {
			if labels["com.openfaas.scale.zero"] == "true" && replicas != 0 {
				scalableFunctions = append(scalableFunctions, f.Name)
			}
		}
	}

	return scalableFunctions, nil
}

func ScaleFunction(replicas int, functionName string) error {

	body, err := json.Marshal(Body{Replicas: replicas})
	if err != nil {
		log.Fatal(err)
	}

	url := gatewayUrl + "/system/scale-function/" + functionName
	_, err = apiRequest(http.MethodPost, url, body)
	if err != nil {
		return err
	}

	return nil
}

func apiRequest(method, url string, body []byte) ([]byte, error) {

	client := &http.Client{
		Timeout: time.Second * 10,
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(user, password)
	response, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	resBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	err = response.Body.Close()
	if err != nil {
		return nil, err
	}

	return resBody, nil
}
