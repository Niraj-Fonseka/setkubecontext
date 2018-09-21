package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"

	yaml "gopkg.in/yaml.v2"
)

type KubernatesConfigView struct {
	APIVersion string `yaml:"apiVersion"`
	Clusters   []struct {
		Cluster struct {
			InsecureSkipTLSVerify bool   `yaml:"insecure-skip-tls-verify"`
			Server                string `yaml:"server"`
		} `yaml:"cluster"`
		Name string `yaml:"name"`
	} `yaml:"clusters"`
	Contexts []struct {
		Context struct {
			Cluster string `yaml:"cluster"`
			User    string `yaml:"user"`
		} `yaml:"context"`
		Name string `yaml:"name"`
	} `yaml:"contexts"`
	CurrentContext      string `yaml:"current-context"`
	ClusterNameList     []string
	SelectedClusterName string
}

func main() {
	kubeConfig := NewKubeConfigView()
	kubeConfig.GetKubectlConfig()
	kubeConfig.InvokeUserPick()
	kubeConfig.SetKubeConfig()
	kubeConfig.PrintCommandsToRun()
}

func NewKubeConfigView() *KubernatesConfigView {
	return &KubernatesConfigView{}
}

func (c *KubernatesConfigView) GetKubectlConfig() {
	cmd := exec.Command("kubectl", "config", "view")

	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}

	if len(strings.Split(out.String(), "kind")[0]) < 0 {
		err = errors.New("Error fetching data from your context")
		if err != nil {
			log.Fatal(err)
		}
	} else {
		err = yaml.Unmarshal([]byte(strings.Split(out.String(), "kind")[0]), c)
		if err != nil {
			log.Fatal(err)
		}
	}

	c.GetListOfServers()
}

func (c *KubernatesConfigView) GetListOfServers() {
	var clusterList []string
	for _, cluster := range c.Contexts {
		clusterList = append(clusterList, cluster.Name)
	}
	c.ClusterNameList = clusterList
}

func (c *KubernatesConfigView) InvokeUserPick() {
	c.PrintListOfClusters()
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter the number of the cluster: ")
	number, err := reader.ReadString('\n')
	if err != nil {
		log.Fatal(err)
	}
	number = strings.TrimSuffix(number, "\n")

	numberToInt, err := strconv.Atoi(number)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(len(c.ClusterNameList))
	if numberToInt < len(c.ClusterNameList) && !(numberToInt > len(c.ClusterNameList)) && numberToInt >= 0 {
		fmt.Println("Picked : ", numberToInt)
		c.SelectedClusterName = c.ClusterNameList[numberToInt]
	} else {
		fmt.Println("Number does not fall in the range of available clusters")
	}
}

func (c *KubernatesConfigView) PrintListOfClusters() {
	for index, clusterName := range c.ClusterNameList {
		fmt.Printf("(%d) %s \n", index, clusterName)
	}
}

func (c *KubernatesConfigView) SetKubeConfig() {
	cmd := exec.Command("kubectl", "config", "use", c.SelectedClusterName)

	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	} else {
		fmt.Println(out.String())
	}

}

func (c *KubernatesConfigView) PrintCommandsToRun() {
	fmt.Printf("Kubernates context has changed from :%s to %s \n", c.CurrentContext, c.SelectedClusterName)
	fmt.Println("Run : kubectl proxy to start the proxy server")
}
