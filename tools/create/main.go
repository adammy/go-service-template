package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"strings"
	"text/template"
)

type fileConfig struct {
	tmpl string
	dir  string
	file string
	data interface{}
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
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			return err
		}
		return err
	}
	return nil
}

func newFileConfig(name, cfgType string) (*fileConfig, error) {
	var (
		tmpl string
		dir  string
		file string
		data map[string]string = make(map[string]string)
	)

	switch cfgType {
	case "api":
		tmpl = "tools/create/templates/openapi.yml.tmpl"
		dir = "api"
		file = fmt.Sprintf("%s.yml", name)
		data["name"] = strings.Title(name)
	case "main":
		tmpl = "tools/create/templates/main.go.tmpl"
		dir = fmt.Sprintf("cmd/%s", name)
		file = "main.go"
		data["name"] = strings.Title(name)
	default:
		return nil, fmt.Errorf("invalid cfg type %s", cfgType)
	}

	return &fileConfig{
		tmpl: tmpl,
		dir:  dir,
		file: file,
		data: data,
	}, nil
}

func createFile(cfg *fileConfig) error {
	var (
		buf bytes.Buffer
	)

	if err := createDir(cfg.dir); err != nil {
		return err
	}

	tmpl := template.Must(template.ParseFiles(cfg.tmpl))
	tmpl.Execute(&buf, cfg.data)

	file, err := os.Create(path.Join(cfg.dir, cfg.file))
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err = file.Write(buf.Bytes()); err != nil {
		return err
	}
	return nil
}

func main() {
	flag.Parse()
	names := flag.Args()
	types := []string{"api", "main"}
	if len(names) >= 1 {
		name := names[0]
		if err := validateName(name); err != nil {
			log.Fatal(err)
		}
		for _, cfgType := range types {
			cfg, err := newFileConfig(name, cfgType)
			if err != nil {
				log.Fatal(err)
			}
			if err := createFile(cfg); err != nil {
				log.Fatalf("failed to create %s service: %s\n", name, err)
			}
		}
	} else {
		log.Fatal("name should not be empty")
	}
}
