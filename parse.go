package main

import (
	"fmt"
	"github.com/distributed-monitoring/policy-engine-sandbox/policyexpr"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
)

var data = `
a: Easy!
b:
  c: 2
  d: [3, 4]
`

var policy_data1 = `
groups:
  - name: test1
    annotation: ["label1", "label2"]
    interval: 10ms //とかなのかなぁ…？
    rules:
      - record: test-rec1
        expr: test-expr1
      - record: test-rec2
        expr: test-expr2
`

type PolicyYaml struct {
	Groups []struct {
		Name       string   `yaml:"name"`
		Annotation []string `yaml:"annotation"`
		Rules      []struct {
			Record string `yaml:"record"`
			Expr   string `yaml:"expr"`
		} `yaml:"rules"`
		Interval     string `yaml:"interval"`
		LastExecuted string // should be time?
	} `yaml:"groups"`
}

func parse_main() PolicyYaml {
	f, err := os.Open("sample.yaml")
	if err != nil {
		fmt.Println("fire open error")
	}
	defer f.Close()

	b, err := ioutil.ReadAll(f)

	p := PolicyYaml{}

	err = yaml.Unmarshal(b, &p)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	fmt.Printf("--- t:\n%v\n\n", p)

	for _, rule := range p.Groups[0].Rules {
		policyexpr.Policyexpr_main(rule.Expr)
	}
	return p
}
