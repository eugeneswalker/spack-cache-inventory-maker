package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"html/template"
	"io/ioutil"
	"os"
	"strings"
)

var (
	inputFile        string
	outputFile       string
	packageOutputDir string
)

type Metadata struct {
	LastMod string `json:"last_mod"`
}

type SpecDetails struct {
	Name         string `json:"name"`
	Version      string `json:"version"`
	Compiler     string `json:"compiler"`
	LastModified string `json:"last_modified_pretty"`
	OS           string `json:"os"`
	Arch         string `json:"arch"`
	SpecFile     string `json:"specfile"`
}

type SpecData struct {
	Name string        `json:"name"`
	Objs []SpecDetails `json:"objs"`
}

type Inventory struct {
	Meta Metadata   `json:"meta"`
	Data []SpecData `json:"data"`
}

type Package struct {
	Name          string
	UppercaseName string
	Tags          string
}

type PageData struct {
	PackageCount  string
	LastUpdatedAt string
	Packages      []Package
}

type Variant struct {
	Name         string
	Compiler     string
	Arch         string
	OS           string
	LastModified string
	SpecFile     string
}

type PackagePage struct {
	Name     string
	Variants []Variant
}

func main() {
	initGlobalOpts()

	b, e := ioutil.ReadFile(inputFile)
	if e != nil {
		fmt.Println("readfile:", e)
		os.Exit(1)
	}

	var inventory Inventory
	if e := json.Unmarshal(b, &inventory); e != nil {
		fmt.Println("unmarshal:", e)
		os.Exit(1)
	}

	packageCount := 0
	for _, d := range inventory.Data {
		for _, _ = range d.Objs {
			packageCount += 1
		}
	}

	var data PageData

	printer := message.NewPrinter(language.English)
	data.PackageCount = printer.Sprintf("%d", packageCount)
	data.LastUpdatedAt = inventory.Meta.LastMod //"11-11-2021 08:17 PST"

	for _, d := range inventory.Data {
		osMap := make(map[string]bool)
		archMap := make(map[string]bool)
		for _, dd := range d.Objs {
			if _, ok := osMap[dd.OS]; !ok && dd.OS != "" {
				osMap[dd.OS] = true
			}
			if _, ok := archMap[dd.Arch]; !ok && dd.Arch != "" {
				archMap[dd.Arch] = true
			}
		}
		var osTags []string
		for k, _ := range osMap {
			osTags = append(osTags, k)
		}

		var archTags []string
		for k, _ := range archMap {
			archTags = append(archTags, k)
		}

		tagList := append(osTags, archTags...)
		tags := strings.Join(tagList, " ")

		p := Package{
			Name:          fmt.Sprintf("%s@%s", d.Name, d.Objs[0].Version),
			UppercaseName: strings.ToUpper(d.Name),
			Tags:          tags,
		}
		data.Packages = append(data.Packages, p)
	}

	f, e := os.Create(outputFile)
	if e != nil {
		fmt.Println(e)
		os.Exit(1)
	}

	tpl, e := template.ParseFiles("index.tpl.html")
	if e != nil {
		fmt.Println("parse:", e)
		os.Exit(1)
	}
	tpl.Execute(f, data)

	vMap := make(map[string]*PackagePage)

	for _, d := range inventory.Data {
		for _, dd := range d.Objs {

			if _, ok := vMap[dd.Name]; !ok {
				vMap[dd.Name] = &PackagePage{}
			}

			v := Variant{
				Name:         fmt.Sprintf("%s@%s", dd.Name, dd.Version),
				Compiler:     dd.Compiler,
				Arch:         dd.Arch,
				OS:           dd.OS,
				LastModified: dd.LastModified,
				SpecFile:     dd.SpecFile,
			}

			vMap[dd.Name].Name = strings.ToUpper(dd.Name)
			vMap[dd.Name].Variants = append(vMap[dd.Name].Variants, v)
		}
	}

	tpl, e = template.ParseFiles("package.tpl.html")
	if e != nil {
		fmt.Println(e)
		os.Exit(1)
	}

	for _, v := range vMap {
		fn := fmt.Sprintf("%s/%s.html", packageOutputDir, v.Name)
		f, e := os.Create(fn)
		if e != nil {
			fmt.Println(e)
			os.Exit(1)
		}
		tpl.Execute(f, v)
		f.Close()
	}
}

func initGlobalOpts() {
	flag.StringVar(&inputFile, "i", "inventory-raw.json", "input file containing inventory data")
	flag.StringVar(&outputFile, "o", "./index.html", "")
	flag.StringVar(&packageOutputDir, "p", "./packages", "package output dir (must exist)")
	flag.Parse()

	fmt.Println("input file =", inputFile)
	fmt.Println("output file =", outputFile)
	fmt.Println("package output dir =", packageOutputDir)
}
