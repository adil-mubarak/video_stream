package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func main() {
	engine := gin.Default()
	engine.Static("/template","./template")

	engine.GET("/stream",videoStream)
	engine.LoadHTMLFiles("./template.index.html")
	engine.GET("/video",func (ctx *gin.Context)  {
			ctx.HTML(http.StatusOK,"index.html",gin.H{})
	})
	engine.Run(":3333")
}

func videoStream(ctx *gin.Context){
	fileName := "video.mp4"
	if !strings.HasSuffix(fileName, ".mp4"){
		ctx.String(http.StatusBadRequest,"video format doesn't support")
		return
	}
	file,err := os.Open("video/"+fileName)
	if err != nil{
		ctx.String(http.StatusNotFound,"video not found")
		return
	}
	defer file.Close()

	stat,err := file.Stat()
	if err != nil{
		ctx.String(http.StatusInternalServerError,"internal error")
		return
	}
	fileSize := stat.Size()
	rangeHeader := ctx.GetHeader("Range")

	if rangeHeader != ""{
		ranges,err := parseRange(rangeHeader,fileSize)
		if err != nil{
			ctx.String(http.StatusBadRequest,err.Error())
			return
		}
		if len(ranges) > 0{
			ctx.Status(http.StatusPartialContent)
			ctx.Header("Content-Range",fmt.Sprintf("bytes %d-%d-%d",ranges[0].start,ranges[0].end,fileSize))
			ctx.Header("Accept-Ranges","bytes")
			ctx.Header("Content-Disposition",fmt.Sprintf("attachment; filename=%s",fileName))
			http.ServeContent(ctx.Writer,ctx.Request,"video/"+fileName,stat.ModTime(),io.NewSectionReader(file,ranges[0].start,ranges[0].end-ranges[0].start+1))
		}
		return
	}
	ctx.Header("Content-Type","video/mp4")
	ctx.Header("Accept-Ranges",fmt.Sprintf("%d",fileSize))
	ctx.File("video/"+fileName)
}

type rangeInfo struct{
	start int64
	end int64
}

func parseRange(rangeHeader string,fileSize int64)([]rangeInfo,error){
	var ranges []rangeInfo
	parts := strings.Split(rangeHeader[6:],"-")
	if len(parts) != 2{
		return nil,fmt.Errorf("invalid range header")
	}
	start,err := strconv.ParseInt(parts[0],10,64)
	if err != nil{
		return nil,fmt.Errorf("invalid ranage header")
	}
	var end int64
	if parts[1] != ""{
		end,err = strconv.ParseInt(parts[1],10,64)
		if err != nil{
			return nil,fmt.Errorf("invalid range header")
		}
	}else{
		end = fileSize - 1
	}
	if start < 0{
		start = 0
	}
	if end > fileSize || end == 0{
		end = fileSize - 1
	}
	ranges = append(ranges, rangeInfo{start: start,end: end})
	return ranges,nil
}