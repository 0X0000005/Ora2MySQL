package converter

// ColumnDef 列定义结构
type ColumnDef struct {
	Name            string // 列名
	DataType        string // 数据类型
	Length          string // 长度（如果适用）
	Precision       string // 精度（用于 NUMBER）
	Scale           string // 小数位数（用于 NUMBER）
	NotNull         bool   // 是否非空
	DefaultValue    string // 默认值
	Comment         string // 列注释
	IsAutoIncrement bool   // 是否自增列（Oracle GENERATED AS IDENTITY）
}

// Constraint 约束定义结构
type Constraint struct {
	Type       string   // 约束类型：PRIMARY KEY, FOREIGN KEY, UNIQUE, CHECK
	Name       string   // 约束名称
	Columns    []string // 涉及的列名
	RefTable   string   // 引用表（外键用）
	RefColumns []string // 引用列（外键用）
	CheckExpr  string   // CHECK 表达式
}

// IndexDef 索引定义结构
type IndexDef struct {
	Name    string   // 索引名称
	Table   string   // 表名
	Columns []string // 索引列
	Unique  bool     // 是否唯一索引
}

// TableDef 表定义结构
type TableDef struct {
	Name        string       // 表名
	Columns     []ColumnDef  // 列定义列表
	Constraints []Constraint // 约束列表
	Comment     string       // 表注释
}

// ViewDef 视图定义结构
type ViewDef struct {
	Name      string   // 视图名
	Columns   []string // 列名列表（可选）
	SelectSQL string   // SELECT 语句
	Comment   string   // 视图注释
}
