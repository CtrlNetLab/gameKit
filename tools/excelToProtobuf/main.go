package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/panjf2000/ants"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"unicode"

	// 因为golang proto没有支持动态从proto文件中解析message，所以这里使用gogo proto
	"github.com/gogo/protobuf/proto"
	"github.com/jhump/protoreflect/desc/protoparse"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/tealeg/xlsx"
)

type ExcelTypeRelation struct {
	ProtoType string
	GoType    reflect.Type
	GoKind    reflect.Kind
	Repeated  bool
}

type CellTypeField struct {
	TypeName  string
	FiledName string
}

const (
	saveProtoDir string = "/configs/"
	saveDataDir  string = "/configs/"
)

var (
	Int32       int32
	String      string
	StringArray []string
	Int32Array  []int32
	Float       float32
	FloatArray  []float32
	Bool        bool
)

var (
	excelPath *string = flag.String("excelPath", "", "excel dir path")
	dataPath  *string = flag.String("dataPath", "", "dir to save output data")
	protoPath *string = flag.String("protoPath", "", "dir to save output proto")
	excelName *string = flag.String("excelName", "", "excelName not found")

	TypeRelationMap map[string]*ExcelTypeRelation = map[string]*ExcelTypeRelation{
		"int":      {"int32", reflect.TypeOf(Int32), reflect.Int32, false},
		"string":   {"string", reflect.TypeOf(String), reflect.String, false},
		"[string]": {"repeated string", reflect.TypeOf(StringArray), reflect.Slice, true},
		"[int]":    {"repeated int32", reflect.TypeOf(Int32Array), reflect.Slice, true},
		"float":    {"float", reflect.TypeOf(Float), reflect.Float32, false},
		"[float]":  {"repeated float", reflect.TypeOf(FloatArray), reflect.Slice, true},
		"bool":     {"bool", reflect.TypeOf(Bool), reflect.Bool, false},
	}

	outProtoPath string
	outDataPath  string
)

func main() {
	flag.Parse()

	if *excelPath == "" || *dataPath == "" || *protoPath == "" {
		fmt.Printf("usage: excelToProtobuf --excelPath xx --dataPath xx --protoPath xx\n")
		return
	}

	outProtoPath = *protoPath + saveProtoDir
	outDataPath = *dataPath + saveDataDir

	_, err := os.Stat(*excelPath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("dir is not exist: %s\n", *excelPath)
			return
		}
	}

	_, err = os.Stat(*protoPath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("proto dir is not exist: %s\n", *protoPath)
			return
		}
	}

	_, err = os.Stat(*dataPath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("data dir is not exist: %s\n", *protoPath)
			return
		}
	}

	_, err = os.Stat(outProtoPath)
	if err != nil {
		if os.IsNotExist(err) {
			os.Mkdir(outProtoPath, 0755)
		}
	}

	_, err = os.Stat(outDataPath)
	if err != nil {
		if os.IsNotExist(err) {
			os.Mkdir(outDataPath, 0755)
		}
	}
	fileList, err := getExcelFileList()
	if err != nil {
		fmt.Printf("get excel files failed. err: %s\n", err.Error())
		return
	}

	var wg sync.WaitGroup
	pool, _ := ants.NewPool(runtime.NumCPU())
	defer pool.Release()

	for _, fileName := range fileList {
		wg.Add(1)
		name := fileName
		if err := pool.Submit(func() {
			defer wg.Done()
			if err := ParseSheet(name); err != nil {
				fmt.Printf("解析文件失败. %s, Err: %s\n", name, err.Error())
			}
		}); err != nil {
			fmt.Printf("submit task failed. err: %s\n", err.Error())
		}
	}
	wg.Wait()

	//files, err := os.ReadDir(*excelPath)
	//if err != nil {
	//	fmt.Printf("打开文件夹失败. Err: %s\n", err.Error())
	//	return
	//}
	//
	//var wg sync.WaitGroup
	//pool, _ := ants.NewPool(runtime.NumCPU())
	//defer pool.Release()
	//for _, file := range files {
	//	extension := filepath.Ext(file.Name())
	//	if extension != ".xlsx" {
	//		continue
	//	}
	//
	//	wg.Add(1)
	//
	//	fileName := file.Name()
	//	if err := pool.Submit(func() {
	//		defer wg.Done()
	//		err := ParseSheet(fileName)
	//		if err != nil {
	//			fmt.Printf("解析文件失败. %s, Err: %s\n", fileName, err.Error())
	//		}
	//	}); err != nil {
	//		fmt.Printf("submit task failed. err: %s\n", err.Error())
	//	}
	//}
	//
	//wg.Wait()
}

// 工作簿名称不支持中文
func ContainsChinese(text string) bool {
	for _, char := range text {
		if unicode.Is(unicode.Scripts["Han"], char) {
			return true
		}
	}
	return false
}

// 检查工作簿名是否合法
func IsSheetNameInvaild(sheetName string) bool {
	if strings.Contains(sheetName, "Sheet") {
		return true
	}

	if ContainsChinese(sheetName) {
		return true
	}

	pattern := "[^a-zA-Z0-9_]"

	regex, err := regexp.Compile(pattern)
	if err != nil {
		return false
	}

	matches := regex.FindAllString(sheetName, -1)
	return len(matches) > 0
}

func ParseSheet(fileName string) error {
	excelFile, err := xlsx.OpenFile(*excelPath + fileName)
	if err != nil {
		return err
	}

	for _, sheet := range excelFile.Sheets {
		if len(sheet.Rows) == 0 || IsSheetNameInvaild(sheet.Name) {
			continue
		}
		protoName := strings.Title(sheet.Name) + "Config"
		fmt.Printf("处理表格:%s, Sheet:%s\n", fileName, sheet.Name)
		protoFilePath := fmt.Sprintf("%s%s.proto", outProtoPath, protoName)
		protoFile, err := os.OpenFile(protoFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			fmt.Printf("写入proto文件失败. Sheet: %s, Err:%s", sheet.Name, err.Error())
			panic(err)
		}

		var sb strings.Builder
		sb.WriteString("/*@brief: ----auto generate----*/\n\n")
		sb.WriteString("syntax = \"proto3\";\n")
		sb.WriteString("option go_package = \"gamekit/pkg/configs;configs\";\n")
		sb.WriteString("package configs;\n\n")
		sb.WriteString(fmt.Sprintf("message %s {\n", protoName))
		filedRows := sheet.Rows[1]
		typeRows := sheet.Rows[2]
		var fieldNum uint32
		for index, fieldCell := range filedRows.Cells {
			if fieldCell.Value == "" || len(typeRows.Cells) < (index+1) || typeRows.Cells[index].Value == "" {
				continue
			}

			fileName := sheet.Rows[1].Cells[index].Value
			fileType := sheet.Rows[2].Cells[index].Value

			if fileName == "" || fileType == "" {
				continue
			}
			if strings.Contains(fileName, "-") {
				oldFileName := fileName
				//如果字段里包含"-"pb 解析会失败这里把-转换成_
				fileName = strings.ReplaceAll(fileName, "-", "_")
				sheet.Rows[1].Cells[index].Value = fileName
				fmt.Printf("替换非法字符串:%s, fileName:%s\n", oldFileName, fileName)
			}

			sb.WriteString(fmt.Sprintf("\t/* \n\t* %s\n\t*/\n", sheet.Rows[0].Cells[index].Value))

			if typeStr, ok := TypeRelationMap[fileType]; !ok {
				fmt.Printf("unkonw type: %s, file: %s, sheet: %s, row: %d, column: %d\n", fileType, fileName, sheet.Name, index, 2)
				goto errRet
			} else {
				sb.WriteString(fmt.Sprintf("\t%s %s = %d;\n", typeStr.ProtoType, fileName, index+1))
				fieldNum++
			}
		}

	errRet:
		sb.WriteString("}\n\n")

		sb.WriteString(fmt.Sprintf("message %sArray {\n", protoName))
		sb.WriteString(fmt.Sprintf("\trepeated %s items = 1;\n}", protoName))
		protoFile.Write([]byte(sb.String()))

		// 处理二进制
		parser := protoparse.Parser{}
		fileDescs, err := parser.ParseFiles(protoFilePath)
		if err != nil {
			fmt.Printf("%s 序列化到文件失败: %s", sheet.Name, err.Error())
			panic(err)
		}

		fileDesc := fileDescs[0]
		protoMessage := fileDesc.FindMessage(fmt.Sprintf("configs.%s", protoName))
		if protoMessage == nil {
			continue
		}

		protoArrayMessage := fileDesc.FindMessage(fmt.Sprintf("configs.%sArray", protoName))
		if protoArrayMessage == nil {
			continue
		}

		typeNameMap := make(map[int]CellTypeField)
		for index, fieldCell := range typeRows.Cells {
			if fieldCell.Value == "" || typeRows.Cells[index].Value == "" {
				continue
			}
			typeNameMap[index] = CellTypeField{TypeName: fieldCell.String(), FiledName: filedRows.Cells[index].String()}
		}

		allMessage := make(map[string][]*dynamic.Message, 0)
		pbMessage := dynamic.NewMessage(protoArrayMessage)
		messages := make([]*dynamic.Message, 0)
		for rowIdex := 3; rowIdex < len(sheet.Rows); rowIdex++ {
			cell := sheet.Rows[rowIdex]
			// 行首一般是主键，如果主键为空则忽略，避免生成空message
			if len(cell.Cells) == 0 || cell.Cells[0].String() == "" {
				continue
			}
			oneMessage := dynamic.NewMessage(protoMessage)
			for columnIndex, column := range cell.Cells {
				typeFiled, ok := typeNameMap[columnIndex]
				if !ok {
					continue
				}

				typeData, ok := TypeRelationMap[typeFiled.TypeName]
				if !ok {
					fmt.Printf("unkonw type: %s, file: %s, sheet: %s, row: %d, column: %d\n", typeFiled.TypeName, fileName, sheet.Name, columnIndex, 2)
					continue
				}
				if typeData.Repeated {
					switch typeData.GoKind {
					case reflect.Slice:
						//fmt.Printf("file: %s, sheet: %s, fieldName:%s\n", file.Name(), sheet.Name, typeFiled.FiledName)
						switch typeData.GoType {
						case reflect.TypeOf(StringArray):
							if column.String() != "" {
								for _, iValue := range strings.Split(column.String(), ",") {
									oneMessage.AddRepeatedFieldByName(typeFiled.FiledName, iValue)
								}
							}
						case reflect.TypeOf(Int32Array):
							for _, iValue := range strings.Split(column.String(), ",") {
								value, err := strconv.ParseInt(iValue, 10, 32)
								if err == nil {
									oneMessage.AddRepeatedFieldByName(typeFiled.FiledName, int32(value))
								}
							}
						case reflect.TypeOf(FloatArray):
							for _, iValue := range strings.Split(column.String(), ",") {
								value, err := strconv.ParseFloat(iValue, 32)
								if err == nil {
									oneMessage.AddRepeatedFieldByName(typeFiled.FiledName, float32(value))
								}
							}
						default:
							fmt.Printf("unknow repeate type. file: %s, sheet: %s, fieldName: %s", fileName, sheet.Name, typeFiled.FiledName)
						}
					default:
						fmt.Printf("unknow repeated go kind. file: %s, sheet: %s, fieldName: %s", fileName, sheet.Name, typeFiled.FiledName)
					}
				} else {
					switch typeData.GoKind {
					case reflect.Bool:
						value, err := strconv.ParseBool(column.String())
						if err == nil {
							oneMessage.SetFieldByName(typeFiled.FiledName, value)
						}
					case reflect.Int32:
						value, err := strconv.ParseInt(column.String(), 10, 32)
						if err == nil {
							oneMessage.SetFieldByName(typeFiled.FiledName, int32(value))
						}
					case reflect.String:
						oneMessage.SetFieldByName(typeFiled.FiledName, column.String())
					case reflect.Float32:
						value, err := strconv.ParseFloat(column.String(), 32)
						if err == nil {
							oneMessage.SetFieldByName(typeFiled.FiledName, float32(value))
						}
					default:
						fmt.Printf("not support type: %s, kind: %v\n", typeFiled.FiledName, typeData.GoKind)
					}
				}
			}

			pbMessage.AddRepeatedFieldByName("items", oneMessage)
			messages = append(messages, oneMessage)
		}
		allMessage["items"] = messages

		jsonFilePath := fmt.Sprintf("%s%s.json", outDataPath, protoName)
		jsonFile, err := os.OpenFile(jsonFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			fmt.Printf("写入proto文件失败. Sheet: %s, Err:%s", sheet.Name, err.Error())
			panic(err)
		}
		defer jsonFile.Close()
		jsonStr, err := json.MarshalIndent(allMessage, "", "  ")
		if err != nil {
			fmt.Printf("%s marshal to json failed. err: %s", sheet.Name, err.Error())
			return err
		}
		_, err = jsonFile.Write(jsonStr)
		if err != nil {
			fmt.Printf("%s write to json failed. err: %s", sheet.Name, err.Error())
		}

		bytesFilePath := fmt.Sprintf("%s%s.bytes", outDataPath, protoName)
		dataFile, err := os.OpenFile(bytesFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			fmt.Printf("写入proto文件失败. Sheet: %s, Err:%s", sheet.Name, err.Error())
			panic(err)
		}
		defer dataFile.Close()
		pbBytes, err := proto.Marshal(pbMessage)
		if err != nil {
			fmt.Printf("%s marshal to proto failed. err: %s", sheet.Name, err.Error())
			panic(err)
		}
		_, err = dataFile.Write(pbBytes)
		if err != nil {
			fmt.Printf("%s write to proto failed. err: %s", sheet.Name, err.Error())
			panic(err)
		}
	}

	return nil
}

func getExcelFileList() ([]string, error) {
	// 如果指定了 excelName，优先使用
	if *excelName != "" {
		names := strings.Split(*excelName, ",")
		var files []string
		for _, name := range names {
			name = strings.TrimSpace(name)
			if !strings.HasSuffix(name, ".xlsx") {
				name += ".xlsx"
			}
			fullPath := filepath.Join(*excelPath, name)
			if _, err := os.Stat(fullPath); err != nil {
				return nil, fmt.Errorf("excel file not exist: %s", fullPath)
			}
			files = append(files, name)
		}
		return files, nil
	}

	// 否则扫描目录
	entries, err := os.ReadDir(*excelPath)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if filepath.Ext(entry.Name()) == ".xlsx" {
			files = append(files, entry.Name())
		}
	}
	return files, nil
}
