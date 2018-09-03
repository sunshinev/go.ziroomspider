package ziroomspider

import (
	"vivi/ziroomspider/ocr"
	"fmt"
	"net/http"
	"github.com/PuerkitoBio/goquery"
	"regexp"
	"encoding/json"
	"github.com/globalsign/mgo"
)

type HouseInfo struct {
	Name    string
	Image   string
	Price   string
	Url     string
	Size    string
	Floor   string
	Room    string
	Loc     []string
	Toilet  int
	Balcony int
}

type RoomPrice struct {
	Image  string
	Offset [][]int
}

func ScanList(url string) bool {
	fmt.Println("地址：", url)
	res, e := http.Get(url);
	if e != nil {
		fmt.Println(">>>>>>>>>>页面报错<<<<<<<<<<<<")
		return false
	}

	defer res.Body.Close()

	doc, e := goquery.NewDocumentFromReader(res.Body)
	if e != nil {
		return false
	}

	// 正则匹配js中的内容寻找价格数据
	// var ROOM_PRICE = {"image":"//static8.ziroom.com/phoenix/pc/images/price/da4554c01a8c0563bf7fc106c3934722s.png","offset":[[7,4,0,2],[7,3,8,2],[7,1,6,2],[7,8,6,2],[7,4,6,2],[7,6,0,2],[7,0,6,2],[7,3,0,2],[7,3,6,2],[7,6,8,2],[7,9,6,2],[7,3,6,2],[7,9,8,2],[7,8,6,2],[7,8,0,2],[7,1,8,2],[7,1,6,2],[7,0,6,2]]};
	//^http://www.flysnow.org/([\d]{4})/([\d]{2})/([\d]{2})/([\w-]+).html$
	regex, _ := regexp.Compile("ROOM_PRICE = (.*);")
	priceInfoRet := regex.FindStringSubmatch(doc.Text())
	priceInfo := priceInfoRet[1]

	// 进行json处理
	priceConfig := RoomPrice{}
	json.Unmarshal([]byte(priceInfo), &priceConfig)

	if priceConfig.Image == "" {
		fmt.Println(">>>>>>>>>>没有更多数据<<<<<<<<<<<<")
		return false
	}

	// 得到图片的地址
	imageUrl := "http:" + priceConfig.Image
	// 解析图片内容
	imageWords := ocr.Parse(imageUrl)
	// 解密规则
	offsetConfig := priceConfig.Offset

	doc.Find("#houseList > li").Each(func(i int, selection *goquery.Selection) {
		// 房屋名称
		houseName := selection.Find("div.txt > h3 > a").Text()
		// 房屋价格
		housePriceSalt := offsetConfig[i]
		housePrice := ""
		for n := range housePriceSalt {
			housePrice = housePrice + string(imageWords[housePriceSalt[n]])
		}
		fmt.Println(houseName, housePrice)
		// 店铺地址
		houseUrl, _ := selection.Find("div.txt > h3 > a").Attr("href")
		// 图片
		houseImage, _ := selection.Find("div.img.pr > a > img").Attr("_src")
		// 面积
		houseSize := selection.Find("div.txt > div > p:nth-child(1) > span:nth-child(1)").Text()
		// 楼层
		houseFloor := selection.Find("div.txt > div > p:nth-child(1) > span:nth-child(2)").Text()
		// 房间
		houseRoom := selection.Find("div.txt > div > p:nth-child(1) > span:nth-child(3)").Text()
		// 解析子页面

		houseDetail, e := scanDetail("http:" + houseUrl)
		if e != nil {
			fmt.Println(e)
		}
		var house = HouseInfo{houseName, "http:" + houseImage, housePrice, "http:" + houseUrl, houseSize, houseFloor, houseRoom, houseDetail.Loc, houseDetail.Toilet, houseDetail.Balcony}

		record(&house)
	})

	return true
}

type HouseDetail struct {
	Loc     []string
	Toilet  int
	Balcony int
}

func scanDetail(url string) (houseDetail *HouseDetail, err error) {
	fmt.Println("地址：", url)
	res, e := http.Get(url);
	if e != nil {
		return nil, fmt.Errorf(">>>>>>>>>>页面报错2<<<<<<<<<<<<")
	}

	defer res.Body.Close()

	doc, e := goquery.NewDocumentFromReader(res.Body)
	if e != nil {
		return nil, e
	}

	houseLat, _ := doc.Find("#mapsearchText").Attr("data-lat")
	houseLng, _ := doc.Find("#mapsearchText").Attr("data-lng")
	houseToilet := doc.Find("body > div.area.clearfix > div.room_detail_right > p > span.toilet").Text()
	houseBalcony := doc.Find("body > div.area.clearfix > div.room_detail_right > p > span.balcony").Text()

	fmt.Println(houseLat, houseLng, isToilet(houseToilet), isBalcony(houseBalcony))

	houseLoc := []string{houseLat, houseLng}

	return &HouseDetail{houseLoc, isToilet(houseToilet), isBalcony(houseBalcony)}, nil
}

/*
独卫
 */
func isToilet(str string) int {
	if str != "" {
		return 1
	}

	return 0
}

/*
独立阳台
 */
func isBalcony(str string) int {
	if str != "" {
		return 1
	}

	return 0
}

/*
写入数据库
 */
func record(recordItem *HouseInfo) error {
	session, e := mgo.Dial("localhost")
	if e != nil {
		return e
	}
	defer session.Close()

	session.SetMode(mgo.Monotonic, true)

	c := session.DB("golang").C("db_ziroom_record")
	e = c.Insert(&recordItem)
	if e != nil {
		return e
	}
	fmt.Println("写入成功")
	return nil
}
