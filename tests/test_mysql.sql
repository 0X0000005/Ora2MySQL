CREATE TABLE departments (
  dept_id BIGINT NOT NULL COMMENT '部门ID',
  dept_name VARCHAR(100) NOT NULL COMMENT '部门名称',
  location VARCHAR(200) COMMENT '部门位置',
  created_date DATETIME DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (dept_id)
) COMMENT='部门信息表' ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE employees (
  emp_id BIGINT NOT NULL COMMENT '员工ID',
  emp_name VARCHAR(100) NOT NULL COMMENT '员工姓名',
  emp_no CHAR(10) NOT NULL COMMENT '员工工号',
  email VARCHAR(100) COMMENT '电子邮箱',
  phone VARCHAR(20),
  hire_date DATETIME DEFAULT CURRENT_TIMESTAMP,
  salary DECIMAL(10,2) COMMENT '基本工资',
  bonus DECIMAL(8,2),
  dept_id BIGINT,
  status TINYINT DEFAULT 1 COMMENT '状态：0-离职，1-在职',
  description LONGTEXT,
  CONSTRAINT pk_emp PRIMARY KEY (emp_id),
  CONSTRAINT uk_emp_no UNIQUE (emp_no),
  CONSTRAINT fk_emp_dept FOREIGN KEY (dept_id) REFERENCES departments (dept_id),
  CONSTRAINT chk_status CHECK (status IN (0,1))
) COMMENT='员工信息表' ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE products (
  product_id BIGINT NOT NULL,
  product_code VARCHAR(50) NOT NULL COMMENT '产品编码',
  product_name VARCHAR(200) NOT NULL COMMENT '产品名称',
  price DECIMAL(12,2) NOT NULL COMMENT '产品价格',
  stock_qty BIGINT DEFAULT 0 COMMENT '库存数量',
  is_active TINYINT DEFAULT 1,
  product_image LONGBLOB,
  product_desc LONGTEXT,
  weight FLOAT,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME,
  PRIMARY KEY (product_id)
) COMMENT='产品信息表' ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE INDEX idx_emp_name ON employees (emp_name);

CREATE UNIQUE INDEX idx_emp_email ON employees (email);

CREATE INDEX idx_emp_dept ON employees (dept_id, hire_date);

