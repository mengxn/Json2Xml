package main

import (
	"bufio"
	"encoding/json"
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

type Owner struct {
	Name string `xml:"itunes:name"`
}

type SubCategory struct {
	Text string `xml:"text,attr"`
}

type Category struct {
	Text     string      `xml:"text,attr"`
	Category SubCategory `xml:"category"`
}

type Image struct {
	Href string `xml:"href,attr"`
}

type Enclosure struct {
	Url    string `xml:"url,attr"`
	Type   string `xml:"type,attr"`
	Length string `xml:"length,attr"`
}

type Item struct {
	Title     string    `xml:"title"`
	Author    string    `xml:"itunes:author"`
	Subtitle  string    `xml:"itunes:subtitle"`
	Summary   string    `xml:"itunes:summary"`
	Image     Image     `xml:"itunes:image"`
	Enclosure Enclosure `xml:"enclosure"`
	Guid      string    `xml:"guid"`
	PubDate   string    `xml:"pubDate"`
	Duration  string    `xml:"itunes:duration"`
}

type Channel struct {
	Copyright   string   `xml:"copyright"`
	Language    string   `xml:"language"`
	Link        string   `xml:"link"`
	Title       string   `xml:"title"`
	Author      string   `xml:"itunes:author"`
	Subtitle    string   `xml:"itunes:subtitle"`
	Summary     string   `xml:"itunes:summary"`
	Owner       Owner    `xml:"itunes:owner"`
	Description string   `xml:"description"`
	Image       Image    `xml:"itunes:image"`
	Category    Category `xml:"itunes:category"`
	Item        []Item   `xml:"item"`
}

type Rss struct {
	Xmlns   string  `xml:"xmlns:itunes,attr"`
	Version string  `xml:"version,attr"`
	Channel Channel `xml:"channel"`
}

type Post struct {
	Title      string `json:"title"`
	Audio      string `json:"audio_url"`
	Image      string `json:"image"`
	Duration   int    `json:"duration"`
	CreateTime string `json:"create_time"`
}

func choose(args ...string) string {
	for _, arg := range args {
		if arg != "" {
			return arg
		}
	}
	return ""
}

func fillCourse(rss *Rss) {
	var temp string
	fmt.Println("fill course info")
	fmt.Println("copyright(default 个人 @iWant.link)")
	fmt.Scanln(&temp)
	rss.Channel.Copyright = choose(temp, "个人 @iWant.link")
	fmt.Println("title")
	fmt.Scanln(&rss.Channel.Title)
	fmt.Println("subtitle(default title)")
	fmt.Scanln(&temp)
	rss.Channel.Subtitle = choose(temp, rss.Channel.Title)
	fmt.Println("language(default zh-ch)")
	fmt.Scanln(&temp)
	rss.Channel.Language = choose(temp, "zh-ch")
	fmt.Println("link")
	fmt.Scanln(&rss.Channel.Link)
	fmt.Println("author")
	fmt.Scanln(&rss.Channel.Author)
	rss.Channel.Owner.Name = rss.Channel.Author
	fmt.Println("desc")
	fmt.Scanln(&rss.Channel.Description)
	rss.Channel.Summary = rss.Channel.Description
	fmt.Println("image")
	fmt.Scanln(&rss.Channel.Image.Href)
	fmt.Println("category")
	fmt.Scanln(&rss.Channel.Category.Text)
	fmt.Println("sub category")
	fmt.Scanln(&rss.Channel.Category.Category.Text)
}

func fillCourseByConf(rss *Rss, conf string) {
	file, err := os.Open(conf)
	if err != nil {
		panic(err.Error())
	}
	defer file.Close()
	config := map[string]string{}
	reader := bufio.NewReader(file)
	for {
		line, _, err := reader.ReadLine()
		if err == io.EOF {
			break
		}
		pair := strings.Split(string(line), "=")
		config[pair[0]] = pair[1]
	}
	rss.Channel = Channel{
		Copyright: config["copyright"],
		Language:  config["language"],
		Link:      config["link"],
		Title:     config["title"],
		Author:    config["author"],
		Subtitle:  config["subtitle"],
		Summary:   config["summary"],
		Owner: Owner{
			Name: config["author"],
		},
		Description: config["description"],
		Image: Image{
			Href: config["image"],
		},
		Category: Category{
			Text: config["category"],
		},
	}
}

func fillContent(rss *Rss, fileName string) {
	bytes, err := ioutil.ReadFile(fileName)
	if err != nil {
		panic(err)
	}
	var posts = make([]Post, 0)
	err = json.Unmarshal(bytes, &posts)
	if err != nil {
		panic(err.Error())
	}

	for _, post := range posts {
		item := Item{
			Title:    post.Title,
			Author:   rss.Channel.Author,
			Subtitle: post.Title,
			Summary:  rss.Channel.Summary,
			Image: Image{
				Href: choose(rss.Channel.Image.Href, post.Image),
			},
			Enclosure: Enclosure{
				Url:    post.Audio,
				Type:   "audio/x-m4a",
				Length: strconv.Itoa(post.Duration),
			},
			Guid:     post.Audio,
			PubDate:  post.CreateTime,
			Duration: strconv.Itoa(post.Duration),
		}

		rss.Channel.Item = append(rss.Channel.Item, item)
	}
}

func createXml(rss *Rss, feedName string) {
	// convert obj to xml
	bytes, err := xml.Marshal(rss)
	if err != nil {
		panic(err.Error())
	}
	// create xml
	file, err := os.Create(feedName)
	if err != nil {
		panic(err.Error())
	}
	// write header
	_, err = fmt.Fprint(file, "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n")
	if err != nil {
		panic(err.Error())
	}
	// write content
	_, err = fmt.Fprint(file, string(bytes))
	if err != nil {
		panic(err.Error())
	}
}

func main() {
	var rss = Rss{
		Xmlns:   "http://www.itunes.com/dtds/podcast-1.0.dtd",
		Version: "2.0",
	}

	var (
		config string
		source string
		target string
	)
	flag.StringVar(&config, "config", "", "config")
	flag.StringVar(&source, "source", "data.json", "source")
	flag.StringVar(&target, "target", "feed.xml", "target")
	flag.Parse()

	if config != "" {
		fillCourseByConf(&rss, config)
	} else {
		fillCourse(&rss)
	}
	fillContent(&rss, source)
	createXml(&rss, target)
}
