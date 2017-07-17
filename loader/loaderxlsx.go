// 基于xlsx的配置装载器
// FIXME 当前版本不支持热更新，因此是readonly

// 可以读取以下的结构的xlsx sheet
// 注释1	    注释2	注释3   	FIXME [不会解析此行]
// 字段1    字段2    字段3
// 值类型1   值类型2  值类型3
// 值       值       值
// 值       值       值
// 值       值       值
package loader

import (
	"errors"
	"io/ioutil"
	"strconv"

	"github.com/gfandada/gserver/logger"
	"github.com/tealeg/xlsx"
)

var (
	dataConfig configs
)

// 行数据
type rows struct {
	records map[string]interface{} // key:行字段名 value:行的数据
}

// 表数据
type table struct {
	data map[uint32]*rows // key:行id（自定义的） value:行的数据集合
	name string           // 表名
}

// 配置数据集
type configs struct {
	tables map[string]*table // key:表名 value:该表的数据集合
}

type Loader struct {
}

func Init(path string) {
	dataConfig.tables = make(map[string]*table)
	dataConfig.init(path)
}

func (c *configs) init(path string) {
	if path == "" {
		logger.Error("loaderxlsx init path is nil")
		return
	}
	dir_list, err := ioutil.ReadDir(path)
	if err != nil {
		logger.Error("loaderxlsx read dir error : %v", err)
		return
	}
	for _, v := range dir_list {
		c.initXlsx(path + v.Name())
	}
}

func (c *configs) initXlsx(name string) {
	xlFile, err := xlsx.OpenFile(name)
	if err != nil {
		logger.Error("loaderxlsx initXlsx %s error : %v", name, err)
		return
	}
	for _, sheet := range xlFile.Sheets {
		// 第1行是字段名
		fileds := sheet.Rows[1]
		// 第2行是字段类型
		filedsType := sheet.Rows[2]
		table := new(table)
		table.name = sheet.Name
		table.data = make(map[uint32]*rows)
		// 接下来是数值
		for i := 3; i < len(sheet.Rows); i++ {
			// 行数据
			rowData := new(rows)
			rowData.records = make(map[string]interface{})
			for j, v := range sheet.Rows[i].Cells {
				c.typeAndField(rowData.records,
					fileds.Cells[j].String(),
					filedsType.Cells[j].String(),
					v)
			}
			// 写表数据
			key, _ := strconv.ParseInt(sheet.Rows[i].Cells[0].String(), 10, 64)
			table.data[uint32(key)] = rowData
		}
		// 写配置
		dataConfig.tables[sheet.Name] = table
	}
}

// 解析字段类型和字段名
func (c *configs) typeAndField(rowData map[string]interface{}, filedName string,
	fieldType string, fieldVlaue *xlsx.Cell) {
	if fieldVlaue == nil || fieldVlaue.String() == "" {
		return
	}
	var value interface{}
	switch fieldType {
	case "uint32":
		ret, err := strconv.ParseInt(fieldVlaue.String(), 10, 64)
		if err != nil {
			logger.Error("typeAndField err filedName %s fieldType %s filedValue %s",
				filedName, fieldType, fieldVlaue.String())
			return
		}
		value = uint32(ret)
	case "int":
		value1, err := fieldVlaue.Int()
		if err != nil {
			logger.Error("typeAndField err filedName %s fieldType %s filedValue %s",
				filedName, fieldType, fieldVlaue.String())
			return
		}
		value = value1
	case "string":
		value = fieldVlaue.String()
	}
	rowData[filedName] = value
}

// 获取配置数据
// @params table 		表名
// @params rowname 		行名
// @params fieldname 	列名
func (l *Loader) Get(table string, row uint32, fieldname string) (interface{}, error) {
	table1, ok := dataConfig.tables[table]
	if !ok {
		return nil, errors.New("table not exist")
	}
	data, ok1 := table1.data[row]
	if !ok1 {
		return nil, errors.New("table.data not exist")
	}
	return data.records[fieldname], nil
}
