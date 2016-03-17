package main

import (
    "strings"
    "gopkg.in/yaml.v2"
    "os"
    "io/ioutil"
)

type Config struct {
    Telegram_token string
    Listen string
    Services []string
    Db struct {
        Sqlite string
    }
}

func (me *Config) getServicesAsString() string {
    return strings.Join(me.Services, "\n")
}

func (me *Config) hasService(service string) bool {
    for _, v := range me.Services {
        if v == service {
            return true
        }
    }
    return false
}

type ConfigReader struct {
    file string
    cfg Config
}

func NewConfigReader(file string) ConfigReader {
    return ConfigReader{file: file}
}

func readContents(f string) ([]byte, error) {
    file, err := os.Open(f)

    if err != nil {
        return nil, err
    }

    fi, err := file.Stat()

    data := make([]byte, fi.Size())
    _, err = file.Read(data)

    if err != nil {
        return nil, err
    }

    return data, nil
}

func (me *ConfigReader) load() (*Config, error) {
    data, err := ioutil.ReadFile(me.file)
    if err != nil {
        return nil, err
    }

    err = yaml.Unmarshal(data, &me.cfg)
    if err != nil {
        return nil, err
    }

    return &me.cfg, nil
}
