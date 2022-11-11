package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

var (
	outputFile  string
	parallelism int
	searchPath  string
)

type TargetSpec struct {
	Name string
}

type ArchSpecV1Json struct {
	Platform   string
	PlatformOS string `json:"platform_os"`
	Target     string
}

type ArchSpecV1 struct {
	Platform   string
	PlatformOS string `yaml:"platform_os"`
	Target     string
}

type ArchSpecV2 struct {
	Platform   string
	PlatformOS string `yaml:"platform_os"`
	Target     TargetSpec
}

func (s *ArchSpecV1) String() string {
	return fmt.Sprintf("%s-%s-%s", s.Platform, s.PlatformOS, s.Target)
}

func (s *ArchSpecV2) String() string {
	return fmt.Sprintf("%s-%s-%s", s.Platform, s.PlatformOS, s.Target.Name)
}

type CompilerSpec struct {
	Name    string
	Version string
}

func (s *CompilerSpec) String() string {
	return fmt.Sprintf("%s@%s", s.Name, s.Version)
}

type Spec struct {
	Name               string `json:"name"`
	Version            string `json:"version"`
	OS                 string `json:"os"`
	Arch               string `json:"arch"`
	Compiler           string `json:"compiler"`
	PackageHash        string `json:"package_hash"`
	Hash               string `json:"hash"`
	FullHash           string `json:"full_hash"`
	BuildHash          string `json:"build_hash"`
	SpecFile           string `json:"specfile"`
	LastModifiedPretty string `json:"last_modified_pretty"`
	LastModified       string `json:"last_modified"`
}

type SpecV1 struct {
	Version     string       `yaml:"version"`
	Arch        ArchSpecV1   `yaml:"arch"`
	Compiler    CompilerSpec `yaml:"compiler"`
	PackageHash string       `yaml:"package_hash"`
	Hash        string       `yaml:"hash"`
	FullHash    string       `yaml:"full_hash"`
	BuildHash   string       `yaml:"build_hash"`
}

type SpecV2 struct {
	Version     string       `yaml:"version"`
	Arch        ArchSpecV2   `yaml:"arch"`
	Compiler    CompilerSpec `yaml:"compiler"`
	PackageHash string       `yaml:"package_hash"`
	Hash        string       `yaml:"hash"`
	FullHash    string       `yaml:"full_hash"`
	BuildHash   string       `yaml:"build_hash"`
}

type SpecV3 struct {
	Name        string         `json:"name"`
	Version     string         `json:"version"`
	Arch        ArchSpecV1Json `json:"arch"`
	Compiler    CompilerSpec   `json:"compiler"`
	PackageHash string         `json:"package_hash"`
	Hash        string         `json:"hash"`
	FullHash    string         `json:"full_hash"`
	BuildHash   string         `json:"build_hash"`
}

type JsonSpec struct {
	Nodes []*SpecV3 `json:"nodes"`
}

type SpecJsonV1 struct {
	Specs *JsonSpec `json:"spec"`
}

type SpecYamlV1 struct {
	Specs []map[string]*SpecV1 `yaml:"spec"`
}

type SpecYamlV2 struct {
	Specs []map[string]*SpecV2 `yaml:"spec"`
}

type SpecYamlV3 struct {
	Specs []map[string]*SpecV3 `yaml:"spec"`
}

func (sj *SpecJsonV1) Spec() Spec {
	s := Spec{}
	n := sj.Specs.Nodes[0]
	s.Name = n.Name
	s.Version = n.Version
	s.Arch = n.Arch.Target
	s.OS = n.Arch.PlatformOS
	s.Compiler = n.Compiler.String()
	s.PackageHash = n.PackageHash
	s.Hash = n.Hash
	s.FullHash = n.FullHash
	s.BuildHash = n.BuildHash
	return s
}

func (sy *SpecYamlV1) Spec() Spec {
	s := Spec{}
	for k := range sy.Specs[0] {
		v1 := sy.Specs[0][k]
		s.Name = k
		s.Version = v1.Version
		s.Arch = v1.Arch.Target
		s.OS = v1.Arch.PlatformOS
		s.Compiler = v1.Compiler.String()
		s.PackageHash = v1.PackageHash
		s.Hash = v1.Hash
		s.FullHash = v1.FullHash
		s.BuildHash = v1.BuildHash
		break
	}
	return s
}

func (sy *SpecYamlV2) Spec() Spec {
	s := Spec{}
	for k := range sy.Specs[0] {
		v := sy.Specs[0][k]
		s.Name = k
		s.Version = v.Version
		s.Arch = v.Arch.Target.Name
		s.OS = v.Arch.PlatformOS
		s.Compiler = v.Compiler.String()
		s.PackageHash = v.PackageHash
		s.Hash = v.Hash
		s.FullHash = v.FullHash
		s.BuildHash = v.BuildHash
		break
	}
	return s
}

func work(paths []string, r chan []Spec, d chan int) {
	fmt.Println("Processing ", len(paths))

	specs := make([]Spec, 0)
	failed := 0

	for i := 0; i < len(paths); i += 1 {

		b, e := ioutil.ReadFile(paths[i])
		if e != nil {
			fmt.Println("error: %s", e)
			os.Exit(1)
		}

		syv1 := SpecYamlV1{}
		syv2 := SpecYamlV2{}

		var spec Spec
		if e = yaml.Unmarshal(b, &syv1); e != nil {
			if e = yaml.Unmarshal(b, &syv2); e != nil {
				failed += 1
				continue
			} else {
				spec = syv2.Spec()
			}
		} else {
			spec = syv1.Spec()
		}

		spec.SpecFile = filepath.Base(paths[i])
		specs = append(specs, spec)
	}

	fmt.Println("# FAILED =", failed)

	r <- specs
}

func workJson(paths []string, r chan []Spec, d chan int) {
	fmt.Println("Processing ", len(paths))

	specs := make([]Spec, 0)
	failed := 0

	for i := 0; i < len(paths); i += 1 {

		b, e := ioutil.ReadFile(paths[i])
		if e != nil {
			fmt.Println("error: %s", e)
			os.Exit(1)
		}

		sjv1 := SpecJsonV1{}

		var spec Spec
		if e := json.Unmarshal(b, &sjv1); e != nil {
			failed += 1
			continue
		} else {
			spec = sjv1.Spec()
		}

		spec.SpecFile = filepath.Base(paths[i])
		specs = append(specs, spec)
	}

	fmt.Println("# FAILED =", failed)

	r <- specs
}

func main() {
	initGlobalOpts()

	specs := make([]Spec, 0)
	done := make(chan int)
	results := make(chan []Spec)

	// .spec.yaml
	fns, err := filesWithSuffix(searchPath, ".spec.yaml")
	if err != nil {
		errf("failed to list .spec.yaml: %s: %v", searchPath, err)
	}

	chunks := chunkify(fns, parallelism)

	fmt.Println("# .spec.yaml = ", len(fns))

	for i := 0; i < parallelism; i += 1 {
		go work(chunks[i], results, done)
	}

	for i := 0; i < parallelism; i += 1 {
		s := <-results
		specs = append(specs, s...)
	}

	// .spec.json
	fns, err = filesWithSuffix(searchPath, ".spec.json")
	if err != nil {
		errf("failed to list .spec.json: %s: %v", searchPath, err)
	}

	chunks = chunkify(fns, parallelism)

	fmt.Println("# .spec.json = ", len(fns))

	for i := 0; i < parallelism; i += 1 {
		go workJson(chunks[i], results, done)
	}

	for i := 0; i < parallelism; i += 1 {
		s := <-results
		specs = append(specs, s...)
	}

	// .spec.json.sig
	fns, err = filesWithSuffix(searchPath, ".spec.json.sig")
	if err != nil {
		errf("failed to list .spec.json.sig: %s: %v", searchPath, err)
	}

	chunks = chunkify(fns, parallelism)

	fmt.Println("# .spec.json.sig = ", len(fns))

	for i := 0; i < parallelism; i += 1 {
		go workJson(chunks[i], results, done)
	}

	for i := 0; i < parallelism; i += 1 {
		s := <-results
		specs = append(specs, s...)
	}

	fmt.Println("# specs total =", len(specs))

	om := make(map[string]Spec)
	for _, s := range specs {
		if _, exists := om[s.SpecFile]; exists {
			fmt.Println("error: should not have found existing key:", s.SpecFile)
			continue
		}

		if s.OS == "" {
			fmt.Println(s)
			os.Exit(1)
		}

		om[s.SpecFile] = s
	}

	j, err := json.MarshalIndent(om, "", " ")
	if err != nil {
		fmt.Println("error: failed to marshal:", err)
		return
	}

	err = ioutil.WriteFile(outputFile, j, 0644)
	if err != nil {
		fmt.Println("error: failed to write file:", err)
		return
	}
}

func initGlobalOpts() {
	flag.StringVar(&outputFile, "o", "inventory-raw.json", "output file")
	flag.IntVar(&parallelism, "n", 1, "degree of parallelism")
	flag.Parse()

	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}

	searchPath = flag.Arg(0)

	fmt.Println("output file =", outputFile)
	fmt.Println("parallelism =", parallelism)
	fmt.Println("path =", searchPath)
}

func filesWithSuffix(d, suffix string) ([]string, error) {
	var fns []string

	err := filepath.Walk(d, func(fn string, f os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if f.IsDir() || !strings.HasSuffix(f.Name(), suffix) {
			return nil
		}

		fns = append(fns, fn)

		return nil
	})

	if err != nil {
		return nil, err
	}

	return fns, nil
}

func chunkify(w []string, n int) [][]string {
	nper := len(w) / n
	if len(w)%n != 0 {
		nper += 1
	}
	var ws [][]string
	lasti := 0
	for i := 0; i < n; i++ {
		nexti := lasti + nper
		if nexti > len(w) || i == n-1 {
			nexti = len(w)
		}
		ws = append(ws, w[lasti:nexti])
		lasti = nexti
	}
	return ws
}

func errf(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, msg+"\n", args...)
	os.Exit(1)
}
