-- 测试用例：包含各种 Oracle DDL 元素

-- 创建部门表
CREATE TABLE departments (
  dept_id NUMBER(10) PRIMARY KEY,
  dept_name VARCHAR2(100) NOT NULL,
  location VARCHAR2(200),
  created_date DATE DEFAULT SYSDATE
);

COMMENT ON TABLE departments IS '部门信息表';
COMMENT ON COLUMN departments.dept_id IS '部门ID';
COMMENT ON COLUMN departments.dept_name IS '部门名称';
COMMENT ON COLUMN departments.location IS '部门位置';

-- 创建员工表
CREATE TABLE employees (
  emp_id NUMBER(10) NOT NULL,
  emp_name VARCHAR2(100) NOT NULL,
  emp_no CHAR(10) NOT NULL,
  email VARCHAR2(100),
  phone VARCHAR2(20),
  hire_date DATE DEFAULT SYSDATE,
  salary NUMBER(10,2),
  bonus NUMBER(8,2),
  dept_id NUMBER(10),
  status NUMBER(1) DEFAULT 1,
  description CLOB,
  CONSTRAINT pk_emp PRIMARY KEY (emp_id),
  CONSTRAINT uk_emp_no UNIQUE (emp_no),
  CONSTRAINT fk_emp_dept FOREIGN KEY (dept_id) REFERENCES departments(dept_id),
  CONSTRAINT chk_status CHECK (status IN (0,1))
);

COMMENT ON TABLE employees IS '员工信息表';
COMMENT ON COLUMN employees.emp_id IS '员工ID';
COMMENT ON COLUMN employees.emp_name IS '员工姓名';
COMMENT ON COLUMN employees.emp_no IS '员工工号';
COMMENT ON COLUMN employees.email IS '电子邮箱';
COMMENT ON COLUMN employees.salary IS '基本工资';
COMMENT ON COLUMN employees.status IS '状态：0-离职，1-在职';

-- 创建索引
CREATE INDEX idx_emp_name ON employees(emp_name);
CREATE UNIQUE INDEX idx_emp_email ON employees(email);
CREATE INDEX idx_emp_dept ON employees(dept_id, hire_date);

-- 创建产品表（测试更多数据类型）
CREATE TABLE products (
  product_id NUMBER(10) PRIMARY KEY,
  product_code VARCHAR2(50) NOT NULL,
  product_name NVARCHAR2(200) NOT NULL,
  price NUMBER(12,2) NOT NULL,
  stock_qty NUMBER(10) DEFAULT 0,
  is_active NUMBER(1) DEFAULT 1,
  product_image BLOB,
  product_desc CLOB,
  weight FLOAT,
  created_at TIMESTAMP DEFAULT SYSTIMESTAMP,
  updated_at TIMESTAMP
);

COMMENT ON TABLE products IS '产品信息表';
COMMENT ON COLUMN products.product_code IS '产品编码';
COMMENT ON COLUMN products.product_name IS '产品名称';
COMMENT ON COLUMN products.price IS '产品价格';
COMMENT ON COLUMN products.stock_qty IS '库存数量';
