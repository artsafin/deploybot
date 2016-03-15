package main

import (
    "strings"
    "gopkg.in/yaml.v2"
    "os"
)

type Config struct {
    Telegram_token string
    Listen string
    Services []string
    Db struct {
        Sqlite string
    }
    access struct {
        Version string
        Tokens []string
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

func (me *Config) hasToken(token string) bool {
    for _, v := range me.access.Tokens {
        if v == token {
            return true
        }
    }
    return false
}

type ConfigReader struct {
    file string
    tokensFile string
    cfg Config
}

func NewConfigReader(file string, tokensFile string) ConfigReader {
    return ConfigReader{file: file, tokensFile: tokensFile}
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
    data, err := readContents(me.file)
    if err != nil {
        return nil, err
    }

    err = yaml.Unmarshal(data, &me.cfg)
    if err != nil {
        return nil, err
    }

    data, err = readContents(me.tokensFile)
    if err != nil {
        return nil, err
    }
    err = yaml.Unmarshal(data, &me.cfg.access)
    if err != nil {
        return nil, err
    }

    return &me.cfg, nil
}
