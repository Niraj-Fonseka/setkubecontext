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

type kubernetesConfigView struct {
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

func NewKubeConfigView() *kubernetesConfigView {
	return &kubernetesConfigView{}
}

func (c *kubernetesConfigView) GetKubectlConfig() {
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

	c.GetClusterList()
}

func (c *kubernetesConfigView) GetClusterList() {
	var clusterList []string
	for _, cluster := range c.Contexts {
		clusterList = append(clusterList, cluster.Name)
	}
	c.ClusterNameList = clusterList
}

func (c *kubernetesConfigView) InvokeUserPick() {
	c.PrintListOfClusters()
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("\nEnter the number of the cluster: ")
	number, err := reader.ReadString('\n')
	if err != nil {
		log.Fatal(err)
	}
	number = strings.TrimSuffix(number, "\n")

	numberToInt, err := strconv.Atoi(number)
	if err != nil {
		log.Fatal(err)
	}

	if numberToInt < len(c.ClusterNameList) && !(numberToInt > len(c.ClusterNameList)) && numberToInt >= 0 {
		c.SelectedClusterName = c.ClusterNameList[numberToInt]
	} else {
		fmt.Println("Number does not fall in the range of available clusters")
	}
}

func (c *kubernetesConfigView) PrintListOfClusters() {
	for index, clusterName := range c.ClusterNameList {
		fmt.Printf("(%d) %s \n", index, clusterName)
	}
}

func (c *kubernetesConfigView) SetKubeConfig() {
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

func (c *kubernetesConfigView) PrintCommandsToRun() {
	fmt.Printf("***  Kubernetes context has changed from :%s to %s ***  \n\n", c.CurrentContext, c.SelectedClusterName)
}
