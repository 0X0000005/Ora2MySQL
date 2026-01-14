package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"o2m/internal/converter"
	"o2m/internal/server"
)

func main() {
	// 定义命令行参数
	inputFile := flag.String("i", "", "输入文件路径（Oracle DDL）")
	outputFile := flag.String("o", "", "输出文件路径（MySQL DDL），不指定则输出到标准输出")
	webMode := flag.Bool("web", false, "启动 Web 服务器模式")
	port := flag.Int("port", 8080, "Web 服务器端口（默认 8080）")
	help := flag.Bool("h", false, "显示帮助信息")

	flag.Parse()

	// 显示帮助信息
	if *help {
		printHelp()
		return
	}

	// Web 服务器模式
	if *webMode {
		err := server.StartWebServer(*port)
		if err != nil {
			fmt.Printf("错误：Web 服务器启动失败: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// 命令行模式
	// 检查输入文件
	if *inputFile == "" {
		fmt.Println("错误：必须指定输入文件或使用 -web 启动 Web 服务器")
		printHelp()
		os.Exit(1)
	}

	// 读取输入文件
	oracleDDL, err := readFile(*inputFile)
	if err != nil {
		fmt.Printf("错误：无法读取输入文件: %v\n", err)
		os.Exit(1)
	}

	// 执行转换
	mysqlDDL, err := converter.ConvertToMySQL(oracleDDL)
	if err != nil {
		fmt.Printf("错误：转换失败: %v\n", err)
		os.Exit(1)
	}

	// 输出结果
	if *outputFile == "" {
		// 输出到标准输出
		fmt.Println(mysqlDDL)
	} else {
		// 写入文件
		err = writeFile(*outputFile, mysqlDDL)
		if err != nil {
			fmt.Printf("错误：无法写入输出文件: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("转换成功！结果已保存到: %s\n", *outputFile)
	}
}

// printHelp 打印帮助信息
func printHelp() {
	fmt.Println("Oracle 转 MySQL DDL 工具")
	fmt.Println()
	fmt.Println("用法:")
	fmt.Println("  命令行模式:")
	fmt.Printf("    %s -i <输入文件> [-o <输出文件>]\n", filepath.Base(os.Args[0]))
	fmt.Println()
	fmt.Println("  Web 服务器模式:")
	fmt.Printf("    %s -web [-port <端口号>]\n", filepath.Base(os.Args[0]))
	fmt.Println()
	fmt.Println("参数:")
	fmt.Println("  -i     输入文件路径（命令行模式必需），包含 Oracle DDL 语句")
	fmt.Println("  -o     输出文件路径（可选），不指定则输出到标准输出")
	fmt.Println("  -web   启动 Web 服务器模式")
	fmt.Println("  -port  Web 服务器端口（默认 8080）")
	fmt.Println("  -h     显示此帮助信息")
	fmt.Println()
	fmt.Println("示例:")
	fmt.Println("  命令行模式:")
	fmt.Printf("    %s -i oracle.sql -o mysql.sql\n", filepath.Base(os.Args[0]))
	fmt.Printf("    %s -i oracle.sql\n", filepath.Base(os.Args[0]))
	fmt.Println()
	fmt.Println("  Web 服务器模式:")
	fmt.Printf("    %s -web\n", filepath.Base(os.Args[0]))
	fmt.Printf("    %s -web -port 9000\n", filepath.Base(os.Args[0]))
	fmt.Println()
	fmt.Println("支持的 Oracle DDL 元素:")
	fmt.Println("  - CREATE TABLE 语句")
	fmt.Println("  - 各种数据类型（VARCHAR2, NUMBER, DATE, CLOB, BLOB 等）")
	fmt.Println("  - 主键、外键、唯一键、检查约束")
	fmt.Println("  - CREATE INDEX 和 CREATE UNIQUE INDEX")
	fmt.Println("  - COMMENT ON TABLE 和 COMMENT ON COLUMN")
}

// readFile 读取文件内容
func readFile(filename string) (string, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// writeFile 写入文件内容
func writeFile(filename, content string) error {
	return os.WriteFile(filename, []byte(content), 0644)
}
