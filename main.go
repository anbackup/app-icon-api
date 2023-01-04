package main

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cache"
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
	app.Use(cache.New(cache.Config{
		Next: func(c *fiber.Ctx) bool {
			return c.Query("refresh") == "true"
		},
		Expiration:   30 * time.Minute,
		CacheControl: true,
	}))

	app.Get("/", func(c *fiber.Ctx) error {
		c.Response().Header.Set("Content-type", "application/json; charset=utf-8")
		return c.JSON(map[string]string{
			"coolapk_icon_url":         "https://icon.0n0.dev/coolapk/{package_name}",
			"qq_icon_url":              "https://icon.0n0.dev/qq/{package_name}",
			"playstore_icon_url":       "https://icon.0n0.dev/playstore/{package_name}",
			"fdroid_icon_url":          "https://icon.0n0.dev/fdroid/{package_name}",
			"icon_image":               "https://icon.0n0.dev/image/{origin}/{package_name}",
			"auto_origin_image_or_url": "https://icon.0n0.dev/{package_name}",
		})
	})
	app.Get("/:packageName", func(c *fiber.Ctx) error {
		s := c.Params("packageName")

		if s2 := getIcon("playstore", s); s2 != "" {
			log.Println("playstore", s2)
			r := client.Get(s2).Do()
			if strings.Contains(r.Status, "200") {
				return c.SendStream(r.Body)
			}
		}
		if s2 := getIcon("coolapk", s); s2 != "" {
			log.Println("coolapk", s2)
			c.Location(s2)
			return c.SendStatus(301)
		}
		if s2 := getIcon("qq", s); s2 != "" {
			log.Println("qq", s2)
			c.Location(s2)
			return c.SendStatus(301)
		}
		if s2 := getIcon("fdroid", s); s2 != "" {
			log.Println("fdroid", s2)
			c.Location(s2)
			return c.SendStatus(301)
		}
		return c.SendFile("default.webp")
	})
	// url
	app.Get("/:origin/:packageName", func(c *fiber.Ctx) error {
		s := c.Params("packageName")
		s1 := c.Params("origin")
		s2 := getIcon(s1, s)
		if s2 != "" {
			log.Println(s2)
			c.Response().Header.Add("location", s2)
			return c.SendStatus(301)
		}
		return c.SendFile("default.webp")
	})
	// image
	app.Get("/image/:origin/:packageName", func(c *fiber.Ctx) error {
		s := c.Params("packageName")
		s1 := c.Params("origin")
		s2 := getIcon(s1, s)
		if s2 != "" {
			log.Println(s2)
			r := client.Get(s2).Do()
			if !strings.Contains(r.Status, "200") {
				return fiber.ErrBadGateway
			}
			return c.SendStream(r.Body)
		}
		return c.SendFile("default.webp")
	})
	app.Listen(":3000")
}

func getIcon(origin string, packageName string) string {
	expr1 := ""
	expr2 := ""
	originUrl := ""
	switch origin {
	case "coolapk":
		{
			originUrl = fmt.Sprintf("https://www.coolapk.com/apk/%s", packageName)
			expr1 = `<div class="apk_topbar">([\s\S]+?)<div class="apk_topba_appinfo">`
			expr2 = `src="(.+?)"`
		}
	case "qq":
		{
			originUrl = fmt.Sprintf("https://sj.qq.com/appdetail/%s", packageName)
			expr1 = `<div class="GameCard([\s\S]+?)</picture>`
			expr2 = `src="(.+?)"`
		}
	case "playstore":
		{
			originUrl = fmt.Sprintf("https://play.google.com/store/apps/details?id=%s", packageName)
			expr1 = `<head>([\s\S]+?)</head>`
			expr2 = `<meta property="og:image" content="(.+?)">`
		}
	case "fdroid":
		{
			originUrl = fmt.Sprintf("https://f-droid.org/packages/%s/", packageName)
			expr1 = `<header class="package-header">([\s\S]+?)</header>`
			expr2 = `src="(.+?)"`
		}
	default:
		return ""
	}
	b := client.Get(originUrl).Do()
	if b.Err != nil || !strings.Contains(b.Status, "200") {
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
