/*
 * @Author: your name
 * @Date: 2019-11-01 18:18:44
 * @LastEditTime: 2020-11-20 18:28:25
 * @LastEditors: Please set LastEditors
 * @Description: In User Settings Edit
 * @FilePath: /test/proto/a.go
 */
package main

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-ini/ini"
	"github.com/gocolly/colly"
)

func main() {

	// fmt.Println("10", 0%0)

	cfg, err := ini.Load("config/app.ini")
	if err != nil {
		log.Fatal("Fail to read file: ", err)
		return
	}
	booksCofs := cfg.Section("books")

	s := 0

	urlstr := booksCofs.Key("url").MustString("https://www.biqiuge.com/book/74937382/")
	fmt.Println("列表页url:", urlstr)
	fileTitle := make([]string, 0)
	filecontent := make([]string, 0)
	c := colly.NewCollector()
	// 超时设定
	c.SetRequestTimeout(100 * time.Second)

	c.UserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/86.0.4240.183 Safari/537.36"

	c.AllowURLRevisit = true
	c.DetectCharset = true
	contentCollector := c.Clone()
	beginRevist := false
	selects := booksCofs.Key("selects").MustString("div[class=listmain]")
	fmt.Println("列表页采集范围:", selects)

	selectURL := booksCofs.Key("selectUrl").MustString("dd>a")
	fmt.Println("列表页采集列表:", selectURL)
	c.OnHTML(selects, func(element *colly.HTMLElement) {
		// fmt.Println("找到内容", element, element.Text)

		element.ForEach(selectURL, func(_ int, eleHref *colly.HTMLElement) {

			// if s > 5 {
			// 	return
			// }

			tmpurl := eleHref.Attr("href")
			//if strings.Index(eleHref.Text,"第一章")!= -1{
			fileTitle = append(fileTitle, eleHref.Text)
			beginRevist = true
			//}
			fmt.Printf("开始请求%s 连接%s", eleHref.Text, tmpurl)
			// 休眠2秒
			// time.Sleep(time.Millisecond * 2)
			if beginRevist {
				chapteurl := eleHref.Request.AbsoluteURL(tmpurl)
				contentCollector.Visit(chapteurl)
			}
			s++
		})

	})
	c.OnRequest(func(r *colly.Request) {
		fmt.Println("visiting", r.URL.String())

		u, err := url.Parse(urlstr)
		if err != nil {
			log.Fatal(err)
		}
		// Request头部设定
		r.Headers.Set("Host", u.Host)
		r.Headers.Set("Connection", "keep-alive")
		r.Headers.Set("Accept", "*/*")
		r.Headers.Set("Origin", u.Host)
		r.Headers.Set("Referer", urlstr)
		r.Headers.Set("Accept-Encoding", "gzip, deflate")
		r.Headers.Set("Accept-Language", "zh-CN, zh;q=0.9")
	})
	contselect := booksCofs.Key("contselect").MustString("div[class=showtxt]")
	fmt.Println("contselect", contselect)

	contentCollector.OnHTML(contselect, func(elementcont *colly.HTMLElement) {
		// fmt.Printf("%s\n", elementcont.Text)
		contentStr := strings.Replace(elementcont.Text, "\t", "", -1)

		contentStr = strings.Replace(contentStr, "   ", "", -1)
		// fmt.Println("contentStr", contentStr)
		if contentStr == "" {
			contentStr = "未找到内容"
		}
		filecontent = append(filecontent, contentStr)
	})
	contentCollector.OnResponse(func(resp *colly.Response) {
		// fmt.Printf("连接下载成功\n", string(resp.Body))
		// fmt.Printf("连接下载成功\n", resp.Request.URL.String())
		fmt.Println("response received", resp.StatusCode)
	})
	c.Visit(urlstr)
	// return
	/*
		fmt.Println(len(filecontent))
		fmt.Println(len(fileTitle))
	*/
	fmt.Println(len(filecontent))
	fmt.Println(len(fileTitle))

	if len(filecontent) < 1 {
		fmt.Println("未采集到内容页面，请检查内容页contselect 设置")
		return
	}

	// num := 3
	// // fileNum := filenum % num

	// var titileStr string

	// fmt.Println("fileTitleA", splitArray(fileTitle, 3))

	strpath, _ := os.Getwd()

	splitFile := booksCofs.Key("splitFile").MustInt(1)
	if splitFile == 1 {

		path := strpath + booksCofs.Key("path").MustString("/files/books.md")
		fmt.Println("path", path)
		err = writeFile(path, fileTitle, filecontent)
		if err != nil {
			fmt.Println(err)
		}
	} else {
		fileTitleArr := splitArray(fileTitle, 3)
		filecontentArr := splitArray(filecontent, 3)

		var path string
		for i := 0; i < len(fileTitleArr); i++ {

			name := "/files/books-" + strconv.Itoa(i+1) + ".md"

			path = strpath + name

			fmt.Println("path", path)
			err = writeFile(path, fileTitleArr[i], filecontentArr[i])
			if err != nil {
				fmt.Println(err)
			}

			// fmt.Println("调用次数", i, path)

		}

	}

}

// 写入文件
func writeFile(path string, fileTitle, filecontent []string) error {

	filenum := len(fileTitle)

	f, err := os.Create(path)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer f.Close()

	for i := 0; i < filenum; i++ {
		//fmt.Println(reflect.TypeOf(fileTitle[i]))

		// fmt.Println("fileTitle[i]", fileTitle)

		titileStr := strings.Replace(fileTitle[i], "\t", "", -1)
		f.WriteString("###" + titileStr + "\r\n")
		f.WriteString(filecontent[i])
		f.WriteString("\r\n")
		f.WriteString("\r\n")

		fmt.Printf("正在写入文件%s,%d\n", titileStr, i)

	}

	return nil

}

//数组平分
func splitArray(arr []string, num int64) [][]string {
	max := int64(len(arr))
	if max < num {
		return nil
	}
	var segmens = make([][]string, 0)
	quantity := max / num
	end := int64(0)
	for i := int64(1); i <= num; i++ {
		qu := i * quantity
		if i != num {
			segmens = append(segmens, arr[i-1+end:qu])
		} else {
			segmens = append(segmens, arr[i-1+end:])
		}
		end = qu - i
	}
	return segmens
}
