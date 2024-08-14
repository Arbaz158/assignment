package main

import (
	"log"

	"github.com/controller"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	//upload excel file and process the data
	r.POST("/upload_excel_and_parse_data", controller.UploadExcelAndFormatData)

	//view
	r.GET("/get_mysq_redis_data", controller.GetMysqlAndRedisData)

	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
