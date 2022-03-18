package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path"
	"strings"
	"text/template"
)

type fileConfig struct {
	templatePath string
	outputDir    string
	outputFile   string
	data         interface{}
}

func validateName(name string) error {
	switch {
	case name == "":
		return errors.New("name should not be empty")
	case strings.Contains(name, "_"):
		return fmt.Errorf("package name %s should not contain underscores", name)
	case strings.Contains(name, "-"):
		return fmt.Errorf("package name %s should not contain hyphens", name)
	case strings.ToLower(name) != name:
		return fmt.Errorf("package name %s should be all lowercase", name)
	default:
		return nil
	}
}

func createDir(dir string) error {
	if _, err := os.Stat(dir); err != nil && errors.Is(err, fs.ErrNotExist) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
		return err
	}
	return nil
}

func newFileConfig(name, cfgType string) (*fileConfig, error) {
	var (
		templatePath string
		outputDir    string
		outputFile   string
		data         map[string]string = make(map[string]string)
	)

	switch cfgType {
	case "api":
		templatePath = "tools/generate-service/templates/openapi.yml.tmpl"
		outputDir = "api"
		outputFile = fmt.Sprintf("%s.yml", name)
		data["name"] = strings.Title(name)
	default:
		return nil, fmt.Errorf("invalid cfg type %s", cfgType)
	}

	return &fileConfig{
		templatePath: templatePath,
		outputDir:    outputDir,
		outputFile:   outputFile,
		data:         data,
	}, nil
}

func createFile(cfg *fileConfig) error {
	var (
		buf bytes.Buffer
	)

	if err := createDir(cfg.outputDir); err != nil {
		return err
	}

	template := template.Must(template.ParseFiles(cfg.templatePath))
	template.Execute(&buf, cfg.data)

	file, err := os.Create(path.Join(cfg.outputDir, cfg.outputFile))
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	_, err = file.Write(buf.Bytes())
	if err != nil {
		log.Fatal(err)
	}
	return nil
}

func main() {
	flag.Parse()
	names := flag.Args()
	if len(names) >= 1 {
		name := names[0]
		if err := validateName(name); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("creating %s service\n", name)
		cfg, err := newFileConfig(name, "api")
		if err != nil {
			log.Fatal(err)
		}
		if err := createFile(cfg); err != nil {
			fmt.Printf("failed to create %s service: %s\n", name, err)
		} else {
			fmt.Printf("finished creating %s service\n", name)
		}
	} else {
		log.Fatal("name should not be empty 1")
	}
}
