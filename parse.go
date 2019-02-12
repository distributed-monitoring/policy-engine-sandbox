package main

import (
	"fmt"
	"log"
	"gopkg.in/yaml.v2"
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
		Name string `yaml:"name"`
		Annotation []string `yaml:"annotation"`
		Rules [] struct {
			Record string `yaml:"record"`
			Expr string `yaml:"expr"`
		} `yaml:"rules"`
		Interval string `yaml:"interval"`
		LastExecuted string // should be time?
	} `yaml:"groups"`
}

func main() {
	p := PolicyYaml{}
    
        err := yaml.Unmarshal([]byte(policy_data1), &p)
        if err != nil {
                log.Fatalf("error: %v", err)
        }
        fmt.Printf("--- t:\n%v\n\n", p)
    
}
