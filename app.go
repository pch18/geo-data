package main

import (
	_ "embed"
	"fmt"
	"log"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
)

//go:embed demo.html
var demoHtml string

//go:embed data/jp.csv
var jpRaw string

//go:embed data/cn.csv
var cnRaw string

var geoData = map[string]map[string]*GeoData{
	"jp": makeGeoData(jpRaw),
	"cn": makeGeoData(cnRaw),
}

var postData = map[string]map[string][]*GeoData{
	"jp": makePostData(geoData["jp"]),
	"cn": makePostData(geoData["cn"]),
}

type GeoNode struct {
	Level    int    `json:"level"`    // 地区级别
	Id       string `json:"id"`       // 地区id
	Parent   string `json:"parent"`   // 上级地区id
	ParentId string `json:"parentId"` // 上级地区id
	PostCode string `json:"postcode"` // 邮政编码
	Name     string `json:"name"`     // 地区名称
	Address  string `json:"address"`  // 地区全称
	Spell    string `json:"spell"`    // 地区拼音或读音
}

type GeoData struct {
	Current  *GeoNode   `json:"current,omitempty"`  // 当前节点
	Parents  []*GeoNode `json:"parents,omitempty"`  // 祖先地区 只存储Parent的副本，不含Parents,Children
	Children []*GeoNode `json:"children,omitempty"` // 子级地区
}

func makePostData(data map[string]*GeoData) map[string][]*GeoData {
	postDataMap := make(map[string][]*GeoData)

	for _, data := range data {
		if data.Current.PostCode != "" {
			postDataMap[data.Current.PostCode] = append(postDataMap[data.Current.PostCode], data)
		}
	}

	return postDataMap
}

func makeGeoData(rawData string) map[string]*GeoData {
	lines := strings.Split(strings.TrimSpace(rawData), "\n")

	geoDataMap := make(map[string]*GeoData)

	var roots []*GeoData
	for _, line := range lines {
		fields := strings.Split(line, ",")
		id := strings.TrimSpace(fields[0])
		parentId := strings.TrimSpace(fields[1])
		name := strings.TrimSpace(fields[2])
		spell := strings.TrimSpace(fields[3])
		address := strings.TrimSpace(fields[4])
		post := strings.TrimSpace(fields[5])

		// 获取或创建当前节点
		cur, exists := geoDataMap[id]
		if !exists {
			cur = &GeoData{
				Current: &GeoNode{Id: id},
			}
			geoDataMap[id] = cur
		}

		// 更新当前节点信息
		cur.Current.Parent = parentId
		cur.Current.ParentId = parentId
		cur.Current.Name = name
		cur.Current.Spell = spell
		cur.Current.Address = address
		cur.Current.PostCode = post

		parentData, parentExists := geoDataMap[parentId]
		if !parentExists {
			parentData = &GeoData{
				Current: &GeoNode{
					Id: parentId,
				},
			}
			geoDataMap[parentId] = parentData
		}

		parentData.Children = append(parentData.Children, cur.Current)

		if parentId == "" {
			roots = append(roots, cur)
		}
	}

	// 构建Parents链和计算Level
	for _, root := range roots {
		buildGeoHierarchy(root, 1, []*GeoNode{}, geoDataMap)
	}
	return geoDataMap
}

func buildGeoHierarchy(cur *GeoData, level int, ancestors []*GeoNode, geoDataMap map[string]*GeoData) {
	if cur == nil {
		return
	}

	// 设置level和parents
	cur.Current.Level = level
	cur.Parents = ancestors

	if len(cur.Children) == 0 {
		return
	}

	// 设置新的newAncestors
	newAncestors := make([]*GeoNode, len(ancestors)+1)
	copy(newAncestors, ancestors)
	newAncestors[len(ancestors)] = cur.Current

	// 递归处理子节点
	for _, c := range cur.Children {
		buildGeoHierarchy(geoDataMap[c.Id], cur.Current.Level+1, newAncestors, geoDataMap)
	}
}

func ginGetGeoData(c *gin.Context) {
	country := c.Param("country")
	id := c.Param("id")

	// 检查国家是否存在
	countryMap, countryExists := geoData[country]
	if !countryExists {
		c.JSON(404, gin.H{"error": "Country not found"})
		return
	}

	// 返回地区或 404
	if info, exists := countryMap[id]; exists {
		c.JSON(200, info)
	} else {
		c.JSON(404, gin.H{"error": "ID not found"})
	}
}

func ginSearchByPost(c *gin.Context) {
	country := c.Param("country")
	postcode := c.Param("postcode")

	// 检查国家是否存在
	countryMap, countryExists := postData[country]
	if !countryExists {
		c.JSON(404, gin.H{"error": "Country not found"})
		return
	}

	// 搜索匹配的邮政编码
	var results = countryMap[postcode]

	if len(results) == 0 {
		c.JSON(404, gin.H{"error": "No matching post code found"})
	} else {
		c.JSON(200, results)
	}
}

func main() {
	r := gin.Default()

	// 添加 gzip 压缩中间件
	r.Use(gzip.Gzip(gzip.DefaultCompression))

	// 添加跨域中间件
	r.Use(cors.Default())

	r.GET("/", func(c *gin.Context) {
		c.Data(200, "text/html; charset=utf-8", []byte(demoHtml))
	})
	// 路由：支持 /:country 和 /:country/:id
	r.GET("/:country", ginGetGeoData)
	r.GET("/:country/:id", ginGetGeoData)
	r.GET("/search_postcode/:country/:postcode", ginSearchByPost)

	// 输出加载的数据统计
	totalCodes := 0
	for countryCode, countryData := range geoData {
		fmt.Printf("Loaded %d %s codes\n", len(countryData), countryCode)
		totalCodes += len(countryData)
	}
	fmt.Printf("Total loaded %d codes\n", totalCodes)

	fmt.Println("Server starting on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
