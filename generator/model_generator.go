package generator

import (
	"fmt"
	"github.com/dave/jennifer/jen"
	"github.com/e421083458/model-generator/database"
	"github.com/e421083458/model-generator/helper"
	"github.com/jinzhu/inflection"
	"github.com/urfave/cli"
	"os"
	"strings"
)

func Generate(c *cli.Context) error {
	addr := "tcp(127.0.0.1:3306)"
	if c.String("a") != "" {
		addr = fmt.Sprintf("tcp(%s)", c.String("a"))
	}
	dbSns := fmt.Sprintf("%s:%s@%s/%s?charset=utf8&parseTime=True&loc=Local",
		c.String("u"), c.String("p"), addr, c.String("d"))
	fmt.Println(dbSns)
	db := database.GetDB(dbSns)
	if c.String("t") == "ALL" {
		tables := db.GetDataBySql("show tables")
		for _, table := range tables {
			tableName := table["Tables_in_"+c.String("d")]
			columns := db.GetDataBySql("desc " + tableName)
			generateModel(tableName, columns, c.String("dir"))
		}
	} else {
		columns := db.GetDataBySql("desc " + c.String("t"))
		generateModel(c.String("t"), columns, c.String("dir"))
	}
	return nil
}

func generateModel(tableName string, columns []map[string]string, dir string) {
	var codes []jen.Code
	for _, col := range columns {
		t := col["Type"]
		column := col["Field"]
		var st *jen.Statement
		if column == "id" {
			st = jen.Id("ID").Uint().Tag(map[string]string{"gorm": "primary_key", "json": "id"})
		} else {
			st = jen.Id(helper.SnakeCase2CamelCase(column, true))
			getCol(st, t)
			st.Tag(map[string]string{"gorm": "column:" + column, "json": column,})
		}
		codes = append(codes, st)
	}
	f := jen.NewFilePath(dir)
	structName := helper.SnakeCase2CamelCase(inflection.Singular(tableName), true)
	f.Type().Id(structName).Struct(codes...)
	_ = os.MkdirAll(dir, os.ModePerm)
	fileName := dir + "/" + inflection.Singular(tableName) + ".go"
	_ = f.Save(fileName)
	fmt.Println(fileName)
}

func getCol(st *jen.Statement, t string) {
	prefix := strings.Split(t, "(")[0]
	switch prefix {
	case "int", "tinyint", "smallint", "mediumint":
		st.Int()
	case "bigint":
		st.Int64()
	case "float":
		st.Float32()
	case "varchar":
		st.String()
	case "decimal":
		st.Float32()
	case "date", "time", "timestamp", "year", "datetime":
		st.Qual("time", "Time")
	default:
		st.String()
	}
}
