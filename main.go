package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/labstack/echo/v4"
)

type Version struct {
	Version    string          `json:"version"`
	Retracted  bool            `json:"retracted"`
	ArchiveUrl string          `json:"archive_url"`
	Pubspec    json.RawMessage `json:"pubspec"`
}

type response struct {
	Name           string     `json:"name"`
	IsDiscontinued bool       `json:"isDiscontinued"`
	ReplacedBy     string     `json:"replacedBy"`
	Latest         Version    `json:"latest"`
	Versions       []*Version `json:"versions"`
}

func updateVersion(pack string, version *Version) {
	version.ArchiveUrl = fmt.Sprintf("http://localhost:8080/dl/%s/%s?real=%s", pack, version.Version, url.QueryEscape(version.ArchiveUrl))
}

func main() {
	e := echo.New()
	e.Debug = true
	e.GET("/api/packages/:package", func(c echo.Context) error {
		e.Logger.Infof("Fetching %s", c.Param("package"))
		resp, err := http.Get(fmt.Sprintf("https://pub.dartlang.org/api/packages/%s", c.Param("package")))
		if err != nil {
			return err
		}
		var r response
		if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
			return err
		}
		updateVersion(r.Name, &r.Latest)
		for _, v := range r.Versions {
			updateVersion(r.Name, v)
		}
		return c.JSON(http.StatusOK, r)
	})
	e.GET("/dl/:package/:version", func(c echo.Context) error {
		e.Logger.Infof("Downloading %s-%s", c.Param("package"), c.Param("version"))
		real, _ := url.QueryUnescape(c.QueryParam("real"))
		resp, err := http.Get(real)
		if err != nil {
			return err
		}
		return c.Stream(resp.StatusCode, resp.Header["Content-Type"][0], resp.Body)
	})
	e.Logger.Fatal(e.Start(":8080"))
}
