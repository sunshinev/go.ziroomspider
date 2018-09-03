
公元2015  第28个秋天

九月的午后，微风吹动窗纱，从24楼看去远处的白云一朵朵的棉花糖浮在空中，两个街角外教堂上的钟敲响了第十三下。

X坐在桌前，双层的书桌上摆满了各种漫画，电脑旁边的《新世纪福音战士》是他最近从旧物箱里重新翻出来的，望了一眼窗外，闭上眼睛深深地吸了一口气。

他又要换房子了，每到这个季节总是要重新换个地方，换一个身份，周围都是陌生人才会让他有安全感，这样就没有人会发现他的秘密。

-----------

打开ZR租房的网站，房源的搜索列表页面映入眼前，适合自己的房子总是自己租不起的。但是X还是要从里面挑选最优的。

Spider是个很不错的选择，Goquery也是个很好的选择


## 军刀工具 Goquery

采用dom的选择器语法，如果使用Chrome非常容易提取元素的选择器。
```
chrome 右键->检查->选择需要的dom元素->代码上右键->copy->copy selector
```

### 安装

```
go get github.com/PuerkitoBio/goquery
```

### 使用方法

读取页面内容生成Document

```
res, e := http.Get(url);
if e != nil {
	// e
}

defer res.Body.Close()

doc, e := goquery.NewDocumentFromReader(res.Body)
if e != nil {
	// e
}

```

使用选择器选择页面内容
```
doc.Find("#houseList > li").Each(func(i int, selection *goquery.Selection) {
	// 房屋名称
	houseName := selection.Find("div.txt > h3 > a").Text()
}
```

或者可以使用直接选取的方式
```
// 获取经纬度
houseLat, _ := doc.Find("#mapsearchText").Attr("data-lat")
houseLng, _ := doc.Find("#mapsearchText").Attr("data-lng")
```


## 反蜘蛛策略

常见的页面元素价格，是类似`<span>$5880</span>`这种实现方式，但是ZR采用了另外一种。
ta的价格是通过CSS样式表对背景图片的偏移来实现的，例如价格`￥2690`的实现：
```
<span style="background-position:1000px" class="num rmb">￥</span>
<span style="background-position:-90px" class="num"></span>
<span style="background-position:-210px" class="num"></span>
<span style="background-position:-0px" class="num"></span>
<span style="background-position:-240px" class="num"></span>
```

图片地址是
```
images/price/0fcc0d83409c547d3a9d038cc7808fa3s.png
```
图片的内容是
```
6532148907
```

那么针对如上的情况X提出一个方案：根据偏移量来转换得到价格

这个思路是对的，但是仔细测试后，发现每次访问，图片的地址都会发生变化，对应的图片里面的数字的排序也会发生变化。

这时X的嘴角露出了微笑，很明显
1. 图片每次访问都发生变化
2. 价格的偏移量数据也会发生变化
3. 为了保证价格每次展示都一样，那么必须有一个跟图片对应的数字转换规则

结论
1. 无法通过直接转换方式
2. 寻找这个图片偏移量和价格的转换规则

于是X通过Chrome的search找到了价格元素修改的js代码
```
var ROOM_PRICE = {"image":"//xxxx.com/phoenix/pc/images/price/0fcc0d83409c547d3a9d038cc7808fa3s.png","offset":[[3,7,0,8],[3,7,0,8],[2,3,0,8],[2,3,7,8],[2,0,2,8],[3,7,2,8],[2,2,7,8],[2,2,2,8],[3,6,7,8],[3,6,7,8],[2,8,2,8],[2,8,7,8],[3,7,0,8],[2,4,7,8],[2,5,7,8],[2,8,7,8],[2,5,7,8],[2,3,7,8]]};


// 这一段不用看了，其实就是将图片上的字符，按照上面ROOM_PRICE的规则，按照数组的索引取出来即可
$('#houseList p.price').each(function(i){
	var dom = $(this);
	if(!ROOM_PRICE['offset'] || !ROOM_PRICE['offset'][i]) return ;
	var pos = ROOM_PRICE['offset'][i];
	for(i in pos){
		var inx = pos.length -i -1;
		var seg = $('<span>', {'style':'background-position:-'+(pos[inx]*offset_unit)+'px', 'class':'num'});
		dom.prepend(seg);
	}
	var seg = $('<span>', {'style':'background-position:1000px', 'class':'num rmb'}).html('￥');
	dom.prepend(seg);
});
```

通过上面这段代码，整个反爬虫策略就暴露无遗了，其实设计者也是心思巧妙。


## OCR 解密之 Tesseract 超立方体

X 在笔记本上写下了如下的话：
1. 我们有了解密规则
2. 有了待解密图片
3. 需要提取图片内容，作为解密的字符串，交由程序处理

于是利用到了另一款军刀工具 `Tesseract`

## 介绍
> Tesseract是一个光学字符识别引擎，支持多种操作系统。Tesseract是基于Apache许可证的自由软件，自2006 年起由Google赞助开发。2006年，Tesseract被认为是最精准的开源光学字符识别引擎之一。

Tesseract 翻译过来是 超立方体

### 安装

这里是Wiki ：https://github.com/tesseract-ocr/tesseract/wiki

Mac 下的安装很简单
```
brew install tesseract

```
安装完毕之后可以解析下试试
```
➜  go tesseract ~/Desktop/7d9a5bb074a89f93a5b4e82bea5dc872s.png stdout --psm 6
2436851907

```

安装Go的package
```
go get -v -t github.com/otiai10/gosseract
```

### 使用方法

直接上代码了，调用很方便

```
package ocr

import (
	"fmt"
	"net/http"
	"os"
	"github.com/otiai10/gosseract"
	"io/ioutil"
	"io"
	"bytes"
)

func Parse(imageUrl string)(string) {

	f, _ := os.Create("s.png")
	defer f.Close()

	resp, _ := http.Get(imageUrl)
	defer  resp.Body.Close()

	pic, _ := ioutil.ReadAll(resp.Body)
	io.Copy(f, bytes.NewReader(pic))

	client := gosseract.NewClient()
	defer client.Close()

	client.SetImage("./s.png")
	text, e := client.Text()

	if e != nil {
		fmt.Println("error")
	}

	return text
}

```

最终，使用OCR解密了图片内容，再通过转换才的到了真正的价格

```
{
    "_id" : ObjectId("5b88dfa8644d03deebc6ba6a"),
    "name" : "天通苑中苑4居室-南卧",
    "image" : "http://img.xxxx.com/pic/house_images/g2m1/M00/66/82/ChAFB1uGM-KAZfN6AAQ6qYhGUtc084.JPG_C_264_198_Q80.jpg",
    "price" : "2290",
    "url" : "http://www.xxxx.com/z/vr/61676366.html",
    "size" : "13 ㎡",
    "floor" : "5/6层",
    "room" : "4室1厅",
    "loc" : [
        "40.077562",
        "116.432684"
    ],
    "toilet" : 0,
    "balcony" : 1
}

/* 2 */
{
    "_id" : ObjectId("5b88dfa9644d03deebc6ba6f"),
    "name" : "天通苑中苑4居室-南卧",
    "image" : "http://img.xxxx.com/pic/house_images/g2m1/M00/4F/6E/ChAFBlt9cXaAY-38AARyYlEU6S4611.JPG_C_264_198_Q80.jpg",
    "price" : "2430",
    "url" : "http://www.xxxx.com/z/vr/61663763.html",
    "size" : "11.8 ㎡",
    "floor" : "5/10层",
    "room" : "4室1厅",
    "loc" : [
        "40.077562",
        "116.432684"
    ],
    "toilet" : 0,
    "balcony" : 1
}

/* 3 */
{
    "_id" : ObjectId("5b88dfa9644d03deebc6ba74"),
    "name" : "天通苑本三区4居室-南卧",
    "image" : "http://img.xxxx.com/pic/house_images/g2m1/M00/5A/E5/ChAFBluBaA-APjRgAASBBLkpT0w953.JPG_C_264_198_Q80.jpg",
    "price" : "2790",
    "url" : "http://www.xxxx.com/z/vr/61666427.html",
    "size" : "25.5 ㎡",
    "floor" : "6/6层",
    "room" : "4室1厅",
    "loc" : [
        "40.066064",
        "116.426734"
    ],
    "toilet" : 0,
    "balcony" : 1
}
```

## END

ZR如果想使用这种方式来做到反爬虫，其实只是限制了一下反制门槛，但是设计很巧妙。其实本次也是初试`Tesseract-OCR`这款军刀工具。












