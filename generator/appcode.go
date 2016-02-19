package generator

import (
	"os"
	"fmt"
	"path"
	"regexp"
	"os/exec"
	"strings"
	"database/sql"
	"path/filepath"

	"github.com/qasico/fire/stubs"
	"github.com/qasico/fire/helper"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
)

const (
	O_MODEL byte = 1 << iota
	O_CONTROLLER
	O_ROUTER
)

type DbTransformer interface {
	GetTableNames(conn *sql.DB) []string
	GetConstraints(conn *sql.DB, table *Table, blackList map[string]bool)
	GetColumns(conn *sql.DB, table *Table, blackList map[string]bool)
	GetGoDataType(sqlType string) string
}

type MysqlDB struct {
}

type PostgresDB struct {
}

var dbDriver = map[string]DbTransformer{
	"mysql":    &MysqlDB{},
	"postgres": &PostgresDB{},
}

type MvcPath struct {
	ModelPath      string
	ControllerPath string
	RouterPath     string
}

var typeMappingMysql = map[string]string{
	"int":                "int", // int signed
	"integer":            "int",
	"tinyint":            "int8",
	"smallint":           "int16",
	"mediumint":          "int32",
	"bigint":             "int64",
	"int unsigned":       "uint", // int unsigned
	"integer unsigned":   "uint",
	"tinyint unsigned":   "uint8",
	"smallint unsigned":  "uint16",
	"mediumint unsigned": "uint32",
	"bigint unsigned":    "uint64",
	"bit":                "uint64",
	"bool":               "bool", // boolean
	"enum":               "string", // enum
	"set":                "string", // set
	"varchar":            "string", // string & text
	"char":               "string",
	"tinytext":           "string",
	"mediumtext":         "string",
	"text":               "string",
	"longtext":           "string",
	"blob":               "string", // blob
	"tinyblob":           "string",
	"mediumblob":         "string",
	"longblob":           "string",
	"date":               "time.Time", // time
	"datetime":           "time.Time",
	"timestamp":          "time.Time",
	"time":               "time.Time",
	"float":              "float32", // float & decimal
	"double":             "float64",
	"decimal":            "float64",
	"binary":             "string", // binary
	"varbinary":          "string",
}

var typeMappingPostgres = map[string]string{
	"serial":                      "int", // serial
	"big serial":                  "int64",
	"smallint":                    "int16", // int
	"integer":                     "int",
	"bigint":                      "int64",
	"boolean":                     "bool", // bool
	"char":                        "string", // string
	"character":                   "string",
	"character varying":           "string",
	"varchar":                     "string",
	"text":                        "string",
	"date":                        "time.Time", // time
	"time":                        "time.Time",
	"timestamp":                   "time.Time",
	"timestamp without time zone": "time.Time",
	"interval":                    "string", // time interval, string for now
	"real":                        "float32", // float & decimal
	"double precision":            "float64",
	"decimal":                     "float64",
	"numeric":                     "float64",
	"money":                       "float64", // money
	"bytea":                       "string", // binary
	"tsvector":                    "string", // fulltext
	"ARRAY":                       "string", // array
	"USER-DEFINED":                "string", // user defined
	"uuid":                        "string", // uuid
	"json":                        "string", // json
}

type Table struct {
	Name          string
	Pk            string
	Uk            []string
	Fk            map[string]*ForeignKey
	Columns       []*Column
	ImportTimePkg bool
}

type Column struct {
	Name string
	Type string
	Tag  *OrmTag
}

type ForeignKey struct {
	Name      string
	RefSchema string
	RefTable  string
	RefColumn string
}

type OrmTag struct {
	Auto        bool
	Pk          bool
	Null        bool
	Index       bool
	Unique      bool
	Column      string
	Size        string
	Decimals    string
	Digits      string
	AutoNow     bool
	AutoNowAdd  bool
	Type        string
	Default     string
	RelOne      bool
	ReverseOne  bool
	RelFk       bool
	ReverseMany bool
	RelM2M      bool
}

func (tb *Table) String() string {
	rv := fmt.Sprintf("type %s struct {\n", camelCase(tb.Name))
	for _, v := range tb.Columns {
		rv += v.String() + "\n"
	}
	rv += "}\n"
	return rv
}

func (col *Column) String() string {
	return fmt.Sprintf("%s %s %s", col.Name, col.Type, col.Tag.String())
}

func (tag *OrmTag) String() string {
	var ormOptions []string
	if tag.Column != "" {
		ormOptions = append(ormOptions, fmt.Sprintf("column(%s)", tag.Column))
	}
	if tag.Auto {
		ormOptions = append(ormOptions, "auto")
	}
	if tag.Size != "" {
		ormOptions = append(ormOptions, fmt.Sprintf("size(%s)", tag.Size))
	}
	if tag.Type != "" {
		ormOptions = append(ormOptions, fmt.Sprintf("type(%s)", tag.Type))
	}
	if tag.Null {
		ormOptions = append(ormOptions, "null")
	}
	if tag.AutoNow {
		ormOptions = append(ormOptions, "auto_now")
	}
	if tag.AutoNowAdd {
		ormOptions = append(ormOptions, "auto_now_add")
	}
	if tag.Decimals != "" {
		ormOptions = append(ormOptions, fmt.Sprintf("digits(%s);decimals(%s)", tag.Digits, tag.Decimals))
	}
	if tag.RelFk {
		ormOptions = append(ormOptions, "rel(fk)")
	}
	if tag.RelOne {
		ormOptions = append(ormOptions, "rel(one)")
	}
	if tag.ReverseOne {
		ormOptions = append(ormOptions, "reverse(one)")
	}
	if tag.ReverseMany {
		ormOptions = append(ormOptions, "reverse(many)")
	}
	if tag.RelM2M {
		ormOptions = append(ormOptions, "rel(m2m)")
	}
	if tag.Pk {
		ormOptions = append(ormOptions, "pk")
	}
	if tag.Unique {
		ormOptions = append(ormOptions, "unique")
	}
	if tag.Default != "" {
		ormOptions = append(ormOptions, fmt.Sprintf("default(%s)", tag.Default))
	}

	if len(ormOptions) == 0 {
		return ""
	}
	return fmt.Sprintf("`orm:\"%s\" json:\"%s\"`", strings.Join(ormOptions, ";"), tag.Column)
}

func GenerateAppcode(driver, connStr, level, tables, currpath string) {
	var mode byte
	switch level {
	case "m":
		mode = O_MODEL
	case "mc":
		mode = O_MODEL | O_CONTROLLER
	case "all":
		mode = O_MODEL | O_CONTROLLER | O_ROUTER
	case "r":
		mode = O_ROUTER
	default:
		helper.ColorLog("[ERRO] Invalid 'level' option: %s\n", level)
		helper.ColorLog("[HINT] Level must be either m, mc or r\n")
		os.Exit(2)
	}
	var selectedTables map[string]bool
	if tables != "" {
		selectedTables = make(map[string]bool)
		for _, v := range strings.Split(tables, ",") {
			selectedTables[v] = true
		}
	}
	switch driver {
	case "mysql":
	case "postgres":
	case "sqlite":
		helper.ColorLog("[ERRO] Generating app code from SQLite database is not supported yet.\n")
		os.Exit(2)
	default:
		helper.ColorLog("[ERRO] Unknown database driver: %s\n", driver)
		helper.ColorLog("[HINT] Driver must be one of mysql, postgres or sqlite\n")
		os.Exit(2)
	}
	gen(driver, connStr, mode, selectedTables, currpath)
}

func gen(dbms, connStr string, mode byte, selectedTableNames map[string]bool, currpath string) {
	db, err := sql.Open(dbms, connStr)
	if err != nil {
		helper.ColorLog("[ERRO] Could not connect to %s database: %s, %s\n", dbms, connStr, err)
		os.Exit(2)
	}
	defer db.Close()
	if trans, ok := dbDriver[dbms]; ok {
		helper.ColorLog("[INFO] Analyzing database tables...\n")
		tableNames := trans.GetTableNames(db)
		tables := getTableObjects(tableNames, db, trans)
		mvcPath := new(MvcPath)
		mvcPath.ModelPath = path.Join(currpath, "models")
		mvcPath.ControllerPath = path.Join(currpath, "controllers")
		mvcPath.RouterPath = path.Join(currpath, "routers")
		createPaths(mode, mvcPath)
		pkgPath := getPackagePath(currpath)
		writeSourceFiles(pkgPath, tables, mode, mvcPath, selectedTableNames)
	} else {
		helper.ColorLog("[ERRO] Generating app code from %s database is not supported yet.\n", dbms)
		os.Exit(2)
	}
}

func (*MysqlDB) GetTableNames(db *sql.DB) (tables []string) {
	rows, err := db.Query("SHOW TABLES")
	if err != nil {
		helper.ColorLog("[ERRO] Could not show tables\n")
		helper.ColorLog("[HINT] Check your connection string\n")
		os.Exit(2)
	}
	defer rows.Close()
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			helper.ColorLog("[ERRO] Could not show tables\n")
			os.Exit(2)
		}
		if name != "migrations" {
			tables = append(tables, name)
		}
	}
	return
}

func getTableObjects(tableNames []string, db *sql.DB, dbTransformer DbTransformer) (tables []*Table) {
	// if a table has a composite pk or doesn't have pk, we can't use it yet
	// these tables will be put into blacklist so that other struct will not
	// reference it.
	blackList := make(map[string]bool)
	// process constraints information for each table, also gather blacklisted table names
	for _, tableName := range tableNames {
		// create a table struct
		tb := new(Table)
		tb.Name = tableName
		tb.Fk = make(map[string]*ForeignKey)
		dbTransformer.GetConstraints(db, tb, blackList)
		tables = append(tables, tb)
	}
	// process columns, ignoring blacklisted tables
	for _, tb := range tables {
		dbTransformer.GetColumns(db, tb, blackList)
	}
	return
}

func (*MysqlDB) GetConstraints(db *sql.DB, table *Table, blackList map[string]bool) {
	rows, err := db.Query(
		`SELECT
			c.constraint_type, u.column_name, u.referenced_table_schema, u.referenced_table_name, referenced_column_name, u.ordinal_position
		FROM
			information_schema.table_constraints c
		INNER JOIN
			information_schema.key_column_usage u ON c.constraint_name = u.constraint_name
		WHERE
			c.table_schema = database() AND c.table_name = ? AND u.table_schema = database() AND u.table_name = ?`,
		table.Name, table.Name) //  u.position_in_unique_constraint,
	if err != nil {
		helper.ColorLog("[ERRO] Could not query INFORMATION_SCHEMA for PK/UK/FK information\n")
		os.Exit(2)
	}
	for rows.Next() {
		var constraintTypeBytes, columnNameBytes, refTableSchemaBytes, refTableNameBytes, refColumnNameBytes, refOrdinalPosBytes []byte
		if err := rows.Scan(&constraintTypeBytes, &columnNameBytes, &refTableSchemaBytes, &refTableNameBytes, &refColumnNameBytes, &refOrdinalPosBytes); err != nil {
			helper.ColorLog("[ERRO] Could not read INFORMATION_SCHEMA for PK/UK/FK information\n")
			os.Exit(2)
		}
		constraintType, columnName, refTableSchema, refTableName, refColumnName, refOrdinalPos :=
		string(constraintTypeBytes), string(columnNameBytes), string(refTableSchemaBytes),
		string(refTableNameBytes), string(refColumnNameBytes), string(refOrdinalPosBytes)
		if constraintType == "PRIMARY KEY" {
			if refOrdinalPos == "1" {
				table.Pk = columnName
			} else {
				table.Pk = ""
				// add table to blacklist so that other struct will not reference it, because we are not
				// registering blacklisted tables
				blackList[table.Name] = true
			}
		} else if constraintType == "UNIQUE" {
			table.Uk = append(table.Uk, columnName)
		} else if constraintType == "FOREIGN KEY" {
			fk := new(ForeignKey)
			fk.Name = columnName
			fk.RefSchema = refTableSchema
			fk.RefTable = refTableName
			fk.RefColumn = refColumnName
			table.Fk[columnName] = fk
		}
	}
}

func (mysqlDB *MysqlDB) GetColumns(db *sql.DB, table *Table, blackList map[string]bool) {
	// retrieve columns
	colDefRows, _ := db.Query(
		`SELECT
			column_name, data_type, column_type, is_nullable, column_default, extra
		FROM
			information_schema.columns
		WHERE
			table_schema = database() AND table_name = ?`,
		table.Name)
	defer colDefRows.Close()
	for colDefRows.Next() {
		// datatype as bytes so that SQL <null> values can be retrieved
		var colNameBytes, dataTypeBytes, columnTypeBytes, isNullableBytes, columnDefaultBytes, extraBytes []byte
		if err := colDefRows.Scan(&colNameBytes, &dataTypeBytes, &columnTypeBytes, &isNullableBytes, &columnDefaultBytes, &extraBytes); err != nil {
			helper.ColorLog("[ERRO] Could not query INFORMATION_SCHEMA for column information\n")
			os.Exit(2)
		}
		colName, dataType, columnType, isNullable, columnDefault, extra :=
		string(colNameBytes), string(dataTypeBytes), string(columnTypeBytes), string(isNullableBytes), string(columnDefaultBytes), string(extraBytes)
		// create a column
		col := new(Column)
		col.Name = camelCase(colName)
		col.Type = mysqlDB.GetGoDataType(dataType)
		// Tag info
		tag := new(OrmTag)
		tag.Column = colName
		if table.Pk == colName {
			col.Name = "Id"
			col.Type = "int"
			if extra == "auto_increment" {
				tag.Auto = true
			} else {
				tag.Pk = true
			}
		} else {
			fkCol, isFk := table.Fk[colName]
			isBl := false
			if isFk {
				_, isBl = blackList[fkCol.RefTable]
			}
			// check if the current column is a foreign key
			if isFk && !isBl {
				tag.RelFk = true
				refStructName := fkCol.RefTable
				col.Name = camelCase(colName)
				col.Type = "*" + camelCase(refStructName)

				if isNullable == "YES" {
					tag.Null = true
				}

			} else {
				// if the name of column is Id, and it's not primary key
				if colName == "id" {
					col.Name = "Id_RENAME"
				}

				if isNullable == "YES" {
					tag.Null = true
				}

				if isSQLSignedIntType(dataType) {
					sign := extractIntSignness(columnType)
					if sign == "unsigned" && extra != "auto_increment" {
						col.Type = mysqlDB.GetGoDataType(dataType + " " + sign)
					}
				}
				if isSQLStringType(dataType) {
					tag.Size = extractColSize(columnType)
				}
				if isSQLTemporalType(dataType) {
					tag.Type = dataType
					//check auto_now, auto_now_add
					if columnDefault == "CURRENT_TIMESTAMP" && extra == "on update CURRENT_TIMESTAMP" {
						tag.AutoNow = true
					}
					// else if columnDefault == "CURRENT_TIMESTAMP" {
					// 	tag.AutoNowAdd = true
					// }
					// need to import time package
					table.ImportTimePkg = true
				}
				if isSQLDecimal(dataType) {
					tag.Digits, tag.Decimals = extractDecimal(columnType)
				}
				if isSQLBinaryType(dataType) {
					tag.Size = extractColSize(columnType)
				}
				if isSQLBitType(dataType) {
					tag.Size = extractColSize(columnType)
				}
			}
		}
		col.Tag = tag
		table.Columns = append(table.Columns, col)
	}
}

func (*MysqlDB) GetGoDataType(sqlType string) (goType string) {
	if v, ok := typeMappingMysql[sqlType]; ok {
		return v
	} else {
		helper.ColorLog("[ERRO] data type (%s) not found!\n", sqlType)
		os.Exit(2)
	}
	return goType
}

func (*PostgresDB) GetTableNames(db *sql.DB) (tables []string) {
	rows, err := db.Query(`
		SELECT table_name FROM information_schema.tables
		WHERE table_catalog = current_database() and table_schema = 'public'`)
	if err != nil {
		helper.ColorLog("[ERRO] Could not show tables: %s\n", err)
		helper.ColorLog("[HINT] Check your connection string\n")
		os.Exit(2)
	}
	defer rows.Close()
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			helper.ColorLog("[ERRO] Could not show tables\n")
			os.Exit(2)
		}
		tables = append(tables, name)
	}
	return
}

func (*PostgresDB) GetConstraints(db *sql.DB, table *Table, blackList map[string]bool) {
	rows, err := db.Query(
		`SELECT
			c.constraint_type,
			u.column_name,
			cu.table_catalog AS referenced_table_catalog,
			cu.table_name AS referenced_table_name,
			cu.column_name AS referenced_column_name,
			u.ordinal_position
		FROM
			information_schema.table_constraints c
		INNER JOIN
			information_schema.key_column_usage u ON c.constraint_name = u.constraint_name
		INNER JOIN
			information_schema.constraint_column_usage cu ON cu.constraint_name =  c.constraint_name
		WHERE
			c.table_catalog = current_database() AND c.table_schema = 'public' AND c.table_name = $1
			AND u.table_catalog = current_database() AND u.table_schema = 'public' AND u.table_name = $2`,
		table.Name, table.Name) //  u.position_in_unique_constraint,
	if err != nil {
		helper.ColorLog("[ERRO] Could not query INFORMATION_SCHEMA for PK/UK/FK information: %s\n", err)
		os.Exit(2)
	}
	for rows.Next() {
		var constraintTypeBytes, columnNameBytes, refTableSchemaBytes, refTableNameBytes, refColumnNameBytes, refOrdinalPosBytes []byte
		if err := rows.Scan(&constraintTypeBytes, &columnNameBytes, &refTableSchemaBytes, &refTableNameBytes, &refColumnNameBytes, &refOrdinalPosBytes); err != nil {
			helper.ColorLog("[ERRO] Could not read INFORMATION_SCHEMA for PK/UK/FK information\n")
			os.Exit(2)
		}
		constraintType, columnName, refTableSchema, refTableName, refColumnName, refOrdinalPos :=
		string(constraintTypeBytes), string(columnNameBytes), string(refTableSchemaBytes),
		string(refTableNameBytes), string(refColumnNameBytes), string(refOrdinalPosBytes)
		if constraintType == "PRIMARY KEY" {
			if refOrdinalPos == "1" {
				table.Pk = columnName
			} else {
				table.Pk = ""
				// add table to blacklist so that other struct will not reference it, because we are not
				// registering blacklisted tables
				blackList[table.Name] = true
			}
		} else if constraintType == "UNIQUE" {
			table.Uk = append(table.Uk, columnName)
		} else if constraintType == "FOREIGN KEY" {
			fk := new(ForeignKey)
			fk.Name = columnName
			fk.RefSchema = refTableSchema
			fk.RefTable = refTableName
			fk.RefColumn = refColumnName
			table.Fk[columnName] = fk
		}
	}
}

func (postgresDB *PostgresDB) GetColumns(db *sql.DB, table *Table, blackList map[string]bool) {
	// retrieve columns
	colDefRows, _ := db.Query(
		`SELECT
			column_name,
			data_type,
			data_type ||
			CASE
				WHEN data_type = 'character' THEN '('||character_maximum_length||')'
				WHEN data_type = 'numeric' THEN '(' || numeric_precision || ',' || numeric_scale ||')'
				ELSE ''
			END AS column_type,
			is_nullable,
			column_default,
			'' AS extra
		FROM
			information_schema.columns
		WHERE
			table_catalog = current_database() AND table_schema = 'public' AND table_name = $1`,
		table.Name)
	defer colDefRows.Close()
	for colDefRows.Next() {
		// datatype as bytes so that SQL <null> values can be retrieved
		var colNameBytes, dataTypeBytes, columnTypeBytes, isNullableBytes, columnDefaultBytes, extraBytes []byte
		if err := colDefRows.Scan(&colNameBytes, &dataTypeBytes, &columnTypeBytes, &isNullableBytes, &columnDefaultBytes, &extraBytes); err != nil {
			helper.ColorLog("[ERRO] Could not query INFORMATION_SCHEMA for column information\n")
			os.Exit(2)
		}
		colName, dataType, columnType, isNullable, columnDefault, extra :=
		string(colNameBytes), string(dataTypeBytes), string(columnTypeBytes), string(isNullableBytes), string(columnDefaultBytes), string(extraBytes)
		// create a column
		col := new(Column)
		col.Name = camelCase(colName)
		col.Type = postgresDB.GetGoDataType(dataType)
		// Tag info
		tag := new(OrmTag)
		tag.Column = colName
		if table.Pk == colName {
			col.Name = "Id"
			col.Type = "int"
			if extra == "auto_increment" {
				tag.Auto = true
			} else {
				tag.Pk = true
			}
		} else {
			fkCol, isFk := table.Fk[colName]
			isBl := false
			if isFk {
				_, isBl = blackList[fkCol.RefTable]
			}
			// check if the current column is a foreign key
			if isFk && !isBl {
				tag.RelFk = true
				refStructName := fkCol.RefTable
				col.Name = camelCase(colName)
				col.Type = "*" + camelCase(refStructName)
			} else {
				// if the name of column is Id, and it's not primary key
				if colName == "id" {
					col.Name = "Id_RENAME"
				}
				if isNullable == "YES" {
					tag.Null = true
				}
				if isSQLStringType(dataType) {
					tag.Size = extractColSize(columnType)
				}
				if isSQLTemporalType(dataType) || strings.HasPrefix(dataType, "timestamp") {
					tag.Type = dataType
					//check auto_now, auto_now_add
					if columnDefault == "CURRENT_TIMESTAMP" && extra == "on update CURRENT_TIMESTAMP" {
						tag.AutoNow = true
					} else if columnDefault == "CURRENT_TIMESTAMP" {
						tag.AutoNowAdd = true
					}
					// need to import time package
					table.ImportTimePkg = true
				}
				if isSQLDecimal(dataType) {
					tag.Digits, tag.Decimals = extractDecimal(columnType)
				}
				if isSQLBinaryType(dataType) {
					tag.Size = extractColSize(columnType)
				}
				if isSQLStrangeType(dataType) {
					tag.Type = dataType
				}
			}
		}
		col.Tag = tag
		table.Columns = append(table.Columns, col)
	}
}

func (*PostgresDB) GetGoDataType(sqlType string) (goType string) {
	if v, ok := typeMappingPostgres[sqlType]; ok {
		return v
	} else {
		helper.ColorLog("[ERRO] data type (%s) not found!\n", sqlType)
		os.Exit(2)
	}
	return goType
}

func createPaths(mode byte, paths *MvcPath) {
	if (mode & O_MODEL) == O_MODEL {
		os.Mkdir(paths.ModelPath, 0777)
	}
	if (mode & O_CONTROLLER) == O_CONTROLLER {
		os.Mkdir(paths.ControllerPath, 0777)
	}
	if (mode & O_ROUTER) == O_ROUTER {
		os.Mkdir(paths.RouterPath, 0777)
	}
}

func writeSourceFiles(pkgPath string, tables []*Table, mode byte, paths *MvcPath, selectedTables map[string]bool) {
	if (O_MODEL & mode) == O_MODEL {
		helper.ColorLog("[INFO] Creating model files...\n")
		writeModelFiles(tables, paths.ModelPath, selectedTables, pkgPath)
	}
	if (O_CONTROLLER & mode) == O_CONTROLLER {
		helper.ColorLog("[INFO] Creating controller files...\n")
		writeControllerFiles(tables, paths.ControllerPath, selectedTables, pkgPath)
	}
	if (O_ROUTER & mode) == O_ROUTER {
		helper.ColorLog("[INFO] Creating router files...\n")
		writeRouterFile(tables, paths.RouterPath, selectedTables, pkgPath)
	}
}

func writeModelFiles(tables []*Table, mPath string, selectedTables map[string]bool, pkgPath string) {
	for _, tb := range tables {
		// if selectedTables map is not nil and this table is not selected, ignore it
		if selectedTables != nil {
			if _, selected := selectedTables[tb.Name]; !selected {
				continue
			}
		}
		filename := getFileName(tb.Name)
		fpath := path.Join(mPath, filename + ".go")
		var f *os.File
		var err error
		if helper.IsExist(fpath) {
			helper.ColorLog("[WARN] %v is exist, do you want to overwrite it? Yes or No?\n", fpath)
			if helper.AskForConfirmation() {
				f, err = os.OpenFile(fpath, os.O_RDWR | os.O_TRUNC, 0666)
				if err != nil {
					helper.ColorLog("[WARN] %v\n", err)
					continue
				}
			} else {
				helper.ColorLog("[WARN] skip create file\n")
				continue
			}
		} else {
			f, err = os.OpenFile(fpath, os.O_CREATE | os.O_RDWR, 0666)
			if err != nil {
				helper.ColorLog("[WARN] %v\n", err)
				continue
			}
		}
		template := ""
		if tb.Pk == "" {
			template = stubs.TemplateModel(false)
		} else {
			template = stubs.TemplateModel(true)
		}

		fileStr := strings.Replace(template, "{{modelStruct}}", tb.String(), 1)
		fileStr = strings.Replace(fileStr, "{{modelName}}", camelCase(tb.Name), -1)
		fileStr = strings.Replace(fileStr, "{{tableName}}", tb.Name, -1)
		timePkg := ""
		importTimePkg := ""

		if tb.ImportTimePkg {
			timePkg = "\"time\"\n"
			importTimePkg = "import \"time\"\n"
		}

		fileStr = strings.Replace(fileStr, "{{pkgPath}}", pkgPath, -1)
		fileStr = strings.Replace(fileStr, "{{timePkg}}", timePkg, -1)
		fileStr = strings.Replace(fileStr, "{{importTimePkg}}", importTimePkg, -1)

		if _, err := f.WriteString(fileStr); err != nil {
			helper.ColorLog("[ERRO] Could not write model file to %s\n", fpath)
			os.Exit(2)
		}
		f.Close()
		helper.ColorLog("[INFO] model => %s\n", fpath)
		formatSourceCode(fpath)
	}
}

func writeControllerFiles(tables []*Table, cPath string, selectedTables map[string]bool, pkgPath string) {
	for _, tb := range tables {
		if selectedTables != nil {
			if _, selected := selectedTables[tb.Name]; !selected {
				continue
			}
		}
		if tb.Pk == "" {
			continue
		}
		filename := getFileName(tb.Name)
		fpath := path.Join(cPath, filename + ".go")
		var f *os.File
		var err error
		if helper.IsExist(fpath) {
			helper.ColorLog("[WARN] %v is exist, do you want to overwrite it? Yes or No?\n", fpath)
			if helper.AskForConfirmation() {
				f, err = os.OpenFile(fpath, os.O_RDWR | os.O_TRUNC, 0666)
				if err != nil {
					helper.ColorLog("[WARN] %v\n", err)
					continue
				}
			} else {
				helper.ColorLog("[WARN] skip create file\n")
				continue
			}
		} else {
			f, err = os.OpenFile(fpath, os.O_CREATE | os.O_RDWR, 0666)
			if err != nil {
				helper.ColorLog("[WARN] %v\n", err)
				continue
			}
		}

		fileStr := strings.Replace(stubs.TemplateController(), "{{ctrlName}}", camelCase(tb.Name), -1)
		fileStr = strings.Replace(fileStr, "{{pkgPath}}", pkgPath, -1)
		if _, err := f.WriteString(fileStr); err != nil {
			helper.ColorLog("[ERRO] Could not write controller file to %s\n", fpath)
			os.Exit(2)
		}
		f.Close()
		helper.ColorLog("[INFO] controller => %s\n", fpath)
		formatSourceCode(fpath)
	}
}

func writeRouterFile(tables []*Table, rPath string, selectedTables map[string]bool, pkgPath string) {
	var nameSpaces []string
	for _, tb := range tables {
		// if selectedTables map is not nil and this table is not selected, ignore it
		if selectedTables != nil {
			if _, selected := selectedTables[tb.Name]; !selected {
				continue
			}
		}
		if tb.Pk == "" {
			continue
		}
		// add name spaces
		tb_name := strings.Replace(tb.Name, "_", "-", -1)
		nameSpace := strings.Replace(stubs.TemplateNamespace(), "{{nameSpace}}", tb_name, -1)
		nameSpace = strings.Replace(nameSpace, "{{ctrlName}}", camelCase(tb.Name), -1)
		nameSpaces = append(nameSpaces, nameSpace)
	}

	// add export controller
	fpath := path.Join(rPath, "router.go")
	routerStr := strings.Replace(stubs.TemplateRouter(), "{{nameSpaces}}", strings.Join(nameSpaces, ""), 1)
	routerStr = strings.Replace(routerStr, "{{pkgPath}}", pkgPath, 1)
	var f *os.File
	var err error
	if helper.IsExist(fpath) {
		helper.ColorLog("[WARN] %v is exist, do you want to overwrite it? Yes or No?\n", fpath)
		if helper.AskForConfirmation() {
			f, err = os.OpenFile(fpath, os.O_RDWR | os.O_TRUNC, 0666)
			if err != nil {
				helper.ColorLog("[WARN] %v\n", err)
				return
			}
		} else {
			helper.ColorLog("[WARN] skip create file\n")
			return
		}
	} else {
		f, err = os.OpenFile(fpath, os.O_CREATE | os.O_RDWR, 0666)
		if err != nil {
			helper.ColorLog("[WARN] %v\n", err)
			return
		}
	}
	if _, err := f.WriteString(routerStr); err != nil {
		helper.ColorLog("[ERRO] Could not write router file to %s\n", fpath)
		os.Exit(2)
	}
	f.Close()
	helper.ColorLog("[INFO] router => %s\n", fpath)
	formatSourceCode(fpath)
}

func formatSourceCode(filename string) {
	cmd := exec.Command("gofmt", "-w", filename)
	if err := cmd.Run(); err != nil {
		helper.ColorLog("[WARN] gofmt err: %s\n", err)
	}
}

func camelCase(in string) string {
	tokens := strings.Split(in, "_")
	for i := range tokens {
		tokens[i] = strings.Title(strings.Trim(tokens[i], " "))
	}
	return strings.Join(tokens, "")
}

func isSQLTemporalType(t string) bool {
	return t == "date" || t == "datetime" || t == "timestamp" || t == "time"
}

func isSQLStringType(t string) bool {
	return t == "char" || t == "varchar"
}

func isSQLSignedIntType(t string) bool {
	return t == "int" || t == "tinyint" || t == "smallint" || t == "mediumint" || t == "bigint"
}

func isSQLDecimal(t string) bool {
	return t == "decimal"
}

func isSQLBinaryType(t string) bool {
	return t == "binary" || t == "varbinary"
}

func isSQLBitType(t string) bool {
	return t == "bit"
}

func isSQLStrangeType(t string) bool {
	return t == "interval" || t == "uuid" || t == "json"
}

func extractColSize(colType string) string {
	regex := regexp.MustCompile(`^[a-z]+\(([0-9]+)\)$`)
	size := regex.FindStringSubmatch(colType)
	return size[1]
}

func extractIntSignness(colType string) string {
	regex := regexp.MustCompile(`(int|smallint|mediumint|bigint)\([0-9]+\)(.*)`)
	signRegex := regex.FindStringSubmatch(colType)
	return strings.Trim(signRegex[2], " ")
}

func extractDecimal(colType string) (digits string, decimals string) {
	decimalRegex := regexp.MustCompile(`decimal\(([0-9]+),([0-9]+)\)`)
	decimal := decimalRegex.FindStringSubmatch(colType)
	digits, decimals = decimal[1], decimal[2]
	return
}

func getFileName(tbName string) (filename string) {
	// avoid test file
	filename = tbName
	for strings.HasSuffix(filename, "_test") {
		pos := strings.LastIndex(filename, "_")
		filename = filename[:pos] + filename[pos + 1:]
	}
	return
}

func getPackagePath(curpath string) (packpath string) {
	gopath := os.Getenv("GOPATH")
	helper.Debugf("gopath:%s", gopath)
	if gopath == "" {
		helper.ColorLog("[ERRO] you should set GOPATH in the env")
		os.Exit(2)
	}

	appsrcpath := ""
	haspath := false
	wgopath := filepath.SplitList(gopath)

	for _, wg := range wgopath {
		wg, _ = filepath.EvalSymlinks(path.Join(wg, "src"))

		if filepath.HasPrefix(strings.ToLower(curpath), strings.ToLower(wg)) {
			haspath = true
			appsrcpath = wg
			break
		}
	}

	if !haspath {
		helper.ColorLog("[ERRO] Can't generate application code outside of GOPATH '%s'\n", gopath)
		os.Exit(2)
	}
	packpath = strings.Join(strings.Split(curpath[len(appsrcpath) + 1:], string(filepath.Separator)), "/")
	return
}
