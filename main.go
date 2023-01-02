package main

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/imroc/req/v3"
)

var client = req.C().
	SetUserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36 Edg/108.0.1462.54").
	SetTimeout(5 * time.Second)

func main() {

	app := fiber.New()
	app.Use(logger.New())
	app.Use(cors.New())

	app.Get("/", func(c *fiber.Ctx) error {
		c.Response().Header.Set("Content-type", "application/json; charset=utf-8")
		return c.JSON(struct {
			Coolapk_icon_url string `json:"coolapk_icon_url"`
			QQ_icon_url      string `json:"qq_icon_url"`
		}{
			Coolapk_icon_url: "https://icon.0n0.dev/coolapk/{package_name}",
			QQ_icon_url:      "https://icon.0n0.dev/qq/{package_name}",
		},
		)
	})
	// coolapk
	app.Get("/:site/:packageName", func(c *fiber.Ctx) error {
		s := c.Params("packageName")
		s1 := c.Params("site")
		s2 := getIcon(s1, s)
		if s2 != "" {
			log.Println(s2)
			c.Response().Header.Add("location", s2)
			return c.SendStatus(301)
		}
		return c.SendFile("default.webp")
	})
	app.Listen(":3000")
}

func getIcon(site string, packageName string) string {
	expr1 := ""
	expr2 := ""
	siteUrl := ""
	switch site {
	case "coolapk":
		{
			siteUrl = fmt.Sprintf("https://www.coolapk.com/apk/%s", packageName)
			expr1 = `<div class="apk_topbar">([\s\S]+?)<div class="apk_topba_appinfo">`
			expr2 = `src="(.+?)"`
			break
		}
	case "qq":
		{
			siteUrl = fmt.Sprintf("https://sj.qq.com/appdetail/%s", packageName)
			expr1 = `<div class="GameCard([\s\S]+?)</picture>`
			expr2 = `src="(.+?)"`
			break
		}
	default:
		return ""
	}
	b := client.Get(siteUrl).Do()
	if !strings.Contains(b.Status, "200") {
		return ""
	}
	r, err := regexp.Compile(expr1)
	if err != nil {
		return ""
	}
	s2 := r.FindString(b.String())
	r, err = regexp.Compile(expr2)
	if err != nil {
		return ""
	}
	s3 := r.FindStringSubmatch(s2)
	if len(s3) > 1 {
		return s3[1]
	}
	return ""
}
