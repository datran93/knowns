package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var (
	dbCache    = make(map[string]*gorm.DB)
	redisCache = make(map[string]*redis.Client)
	cacheMutex sync.RWMutex
)

func getGormDB(connStr string) (*gorm.DB, error) {
	cacheMutex.RLock()
	if db, ok := dbCache[connStr]; ok {
		cacheMutex.RUnlock()
		sqlDB, err := db.DB()
		if err == nil && sqlDB.Ping() == nil {
			return db, nil
		}
	}
	cacheMutex.RUnlock()

	var dialector gorm.Dialector

	if strings.HasPrefix(connStr, "postgresql://") || strings.HasPrefix(connStr, "postgres://") {
		dialector = postgres.Open(connStr)
	} else if strings.HasPrefix(connStr, "mysql://") {
		dialector = mysql.Open(connStr)
	} else if strings.HasPrefix(connStr, "sqlite://") {
		rawPath := strings.TrimPrefix(connStr, "sqlite://")
		if parsed, parseErr := url.Parse(connStr); parseErr == nil {
			rawPath = parsed.Host + parsed.Path
		}
		if rawPath == "" {
			return nil, fmt.Errorf("sqlite connection string is missing the file path")
		}
		dialector = sqlite.Open(rawPath)
	}

	db, err := gorm.Open(dialector, &gorm.Config{DisableAutomaticPing: false})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err == nil {
		sqlDB.SetMaxOpenConns(5)
		sqlDB.SetConnMaxLifetime(time.Hour)
	}

	cacheMutex.Lock()
	dbCache[connStr] = db
	cacheMutex.Unlock()

	return db, nil
}

func getRedisClient(connStr string) (*redis.Client, error) {
	cacheMutex.RLock()
	if client, ok := redisCache[connStr]; ok {
		cacheMutex.RUnlock()
		return client, nil
	}
	cacheMutex.RUnlock()

	opts, err := redis.ParseURL(connStr)
	if err != nil {
		return nil, err
	}

	client := redis.NewClient(opts)

	cacheMutex.Lock()
	redisCache[connStr] = client
	cacheMutex.Unlock()

	return client, nil
}

func isRedis(connStr string) bool {
	return strings.HasPrefix(connStr, "redis://") || strings.HasPrefix(connStr, "rediss://")
}

const (
	denyRegex = `(?i)\b(INSERT|UPDATE|DELETE|DROP|ALTER|TRUNCATE|GRANT|REVOKE|CREATE)\b`
	maxRows   = 2000
)

var writeDenyRegex = regexp.MustCompile(denyRegex)

func safeRegexMatch(query string) bool {
	return writeDenyRegex.MatchString(query)
}

// RegisterDatabaseTool registers the database inspector MCP tool.
func RegisterDatabaseTool(s *server.MCPServer) {
	s.AddTool(
		mcp.NewTool("database",
			mcp.WithDescription("Database operations for SQL (Postgres, MySQL, SQLite) and Redis. Use 'action' to specify: list_tables, inspect_schema, run_query, redis_*."),
			mcp.WithString("action",
				mcp.Required(),
				mcp.Description("Action to perform"),
				mcp.Enum("list_tables", "inspect_schema", "run_query", "redis_list_keys", "redis_inspect_key", "redis_run_command"),
			),
			mcp.WithString("connection_string",
				mcp.Description("Database connection string (postgresql://, mysql://, sqlite://, redis://)"),
			),
			mcp.WithString("table",
				mcp.Description("Table name for schema inspection or sampling"),
			),
			mcp.WithString("query",
				mcp.Description("SQL query to execute or Redis command"),
			),
			mcp.WithNumber("limit",
				mcp.Description("Row limit for query results (default: 100, max: 2000)"),
			),
			mcp.WithNumber("offset",
				mcp.Description("Query offset for pagination (default: 0)"),
			),
			mcp.WithBoolean("confirm",
				mcp.Description("Confirm write operations (required for INSERT/UPDATE/DELETE)"),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			action, err := req.RequireString("action")
			if err != nil {
				return errResult("action is required")
			}
			args := req.GetArguments()
			connStr, _ := args["connection_string"].(string)

			if connStr == "" {
				connStr = os.Getenv("DATABASE_CONNECTION_STRING")
			}
			if connStr == "" && !strings.HasPrefix(action, "redis_") {
				return errResult("connection_string is required")
			}

			switch action {
			case "list_tables":
				return handleDBListTables(connStr)
			case "inspect_schema":
				table, _ := args["table"].(string)
				return handleDBInspectSchema(connStr, table)
			case "run_query":
				query, _ := args["query"].(string)
				limit := 100
				if v, ok := args["limit"].(float64); ok {
					limit = int(v)
				}
				offset := 0
				if v, ok := args["offset"].(float64); ok {
					offset = int(v)
				}
				confirm := false
				if v, ok := args["confirm"].(bool); ok {
					confirm = v
				}
				return handleDBRunQuery(connStr, query, limit, offset, confirm)
			case "redis_list_keys":
				return handleRedisListKeys(connStr)
			case "redis_inspect_key":
				key, _ := args["table"].(string)
				return handleRedisInspectKey(connStr, key)
			case "redis_run_command":
				query, _ := args["query"].(string)
				confirm := false
				if v, ok := args["confirm"].(bool); ok {
					confirm = v
				}
				return handleRedisRunCommand(connStr, query, confirm)
			default:
				return errResultf("unknown database action: %s", action)
			}
		},
	)
}

func handleDBListTables(connStr string) (*mcp.CallToolResult, error) {
	if isRedis(connStr) {
		return handleRedisListKeys(connStr)
	}

	db, err := getGormDB(connStr)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error connecting to db: %v", err)), nil
	}

	tables, err := db.Migrator().GetTables()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error getting tables: %v", err)), nil
	}

	if len(tables) == 0 {
		return mcp.NewToolResultText("No tables found in this database."), nil
	}

	var result strings.Builder
	result.WriteString("Tables:\n")
	for _, t := range tables {
		result.WriteString(fmt.Sprintf("- %s\n", t))
	}
	return mcp.NewToolResultText(result.String()), nil
}

func handleDBInspectSchema(connStr string, tableName string) (*mcp.CallToolResult, error) {
	if isRedis(connStr) {
		return handleRedisInspectKey(connStr, tableName)
	}

	if tableName == "" {
		return errResult("table name is required")
	}

	db, err := getGormDB(connStr)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error connecting to db: %v", err)), nil
	}

	migrator := db.Migrator()
	if !migrator.HasTable(tableName) {
		return mcp.NewToolResultText(fmt.Sprintf("Table '%s' does not exist.", tableName)), nil
	}

	colTypes, err := migrator.ColumnTypes(tableName)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error getting columns: %v", err)), nil
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Schema for table: %s\n\n", tableName))
	result.WriteString("Columns:\n")

	for _, col := range colTypes {
		colType, _ := col.ColumnType()
		nullable := "NOT NULL"
		if nullRef, ok := col.Nullable(); ok && nullRef {
			nullable = "NULL"
		}
		prefix := "  "
		if pk, ok := col.PrimaryKey(); ok && pk {
			prefix = "PK"
		}
		result.WriteString(fmt.Sprintf("%s %s : %s (%s)\n", prefix, col.Name(), colType, nullable))
	}

	return mcp.NewToolResultText(result.String()), nil
}

func handleDBRunQuery(connStr string, query string, limit int, offset int, confirm bool) (*mcp.CallToolResult, error) {
	if query == "" {
		return errResult("query is required")
	}

	if safeRegexMatch(query) && !confirm {
		upperQuery := strings.ToUpper(strings.TrimSpace(query))
		if !strings.HasPrefix(upperQuery, "SELECT") {
			return mcp.NewToolResultText(fmt.Sprintf("SECURITY BLOCK: Only SELECT queries allowed without confirmation.\nQuery: %s\nSet confirm=true to execute.", query)), nil
		}
	}

	if limit > maxRows {
		limit = maxRows
	}

	db, err := getGormDB(connStr)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error connecting to db: %v", err)), nil
	}

	cleanQuery := strings.TrimSpace(query)
	upperQuery := strings.ToUpper(cleanQuery)
	if !strings.Contains(upperQuery, "LIMIT") && strings.HasPrefix(upperQuery, "SELECT") {
		cleanQuery = fmt.Sprintf("%s LIMIT %d OFFSET %d", cleanQuery, limit, offset)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	rows, err := sqlDB.Query(cleanQuery)
	if err != nil {
		return mcp.NewToolResultText(fmt.Sprintf("Error: %v", err)), nil
	}
	defer rows.Close()

	cols, _ := rows.Columns()
	var allRows []map[string]interface{}
	for rows.Next() {
		columns := make([]interface{}, len(cols))
		columnPointers := make([]interface{}, len(cols))
		for i := range columns {
			columnPointers[i] = &columns[i]
		}
		if err := rows.Scan(columnPointers...); err == nil {
			rowMap := make(map[string]interface{})
			for i, colName := range cols {
				val := columnPointers[i].(*interface{})
				if b, ok := (*val).([]byte); ok {
					rowMap[colName] = string(b)
				} else if t, ok := (*val).(time.Time); ok {
					rowMap[colName] = t.Format(time.RFC3339)
				} else {
					rowMap[colName] = *val
				}
			}
			allRows = append(allRows, rowMap)
		}
	}

	if len(allRows) == 0 {
		return mcp.NewToolResultText("Query executed. 0 rows returned."), nil
	}

	jsonBytes, err := json.MarshalIndent(allRows, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	suffix := fmt.Sprintf("\n(Showing %d rows starting at offset %d)", len(allRows), offset)
	return mcp.NewToolResultText(fmt.Sprintf("Query Results:\n```json\n%s\n```%s", string(jsonBytes), suffix)), nil
}

func handleRedisListKeys(connStr string) (*mcp.CallToolResult, error) {
	client, err := getRedisClient(connStr)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error connecting to redis: %v", err)), nil
	}

	ctx := context.Background()
	dbSize, err := client.DBSize(ctx).Result()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	var cursor uint64
	keys, _, err := client.Scan(ctx, cursor, "*", 100).Result()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Redis Database (Size: %d keys)\n\nSample Keys (up to 100):\n", dbSize))
	for _, k := range keys {
		result.WriteString(fmt.Sprintf("- %s\n", k))
	}

	if len(keys) == 0 {
		return mcp.NewToolResultText("No keys found in this Redis database."), nil
	}

	return mcp.NewToolResultText(result.String()), nil
}

func handleRedisInspectKey(connStr string, keyName string) (*mcp.CallToolResult, error) {
	if keyName == "" {
		return errResult("key name is required")
	}

	client, err := getRedisClient(connStr)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error connecting to redis: %v", err)), nil
	}

	ctx := context.Background()
	exists, err := client.Exists(ctx, keyName).Result()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	if exists == 0 {
		return mcp.NewToolResultText(fmt.Sprintf("Key '%s' does not exist in Redis.", keyName)), nil
	}

	keyType, _ := client.Type(ctx, keyName).Result()
	ttl, _ := client.TTL(ctx, keyName).Result()

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Schema for key: %s\n\n", keyName))
	result.WriteString(fmt.Sprintf("Type: %s\n", keyType))
	result.WriteString(fmt.Sprintf("TTL: %v seconds\n\n", ttl.Seconds()))

	switch keyType {
	case "string":
		valLen, _ := client.StrLen(ctx, keyName).Result()
		result.WriteString(fmt.Sprintf("String Length: %d\n", valLen))
	case "hash":
		hlen, _ := client.HLen(ctx, keyName).Result()
		result.WriteString(fmt.Sprintf("Hash Entries: %d\n", hlen))
	case "list":
		llen, _ := client.LLen(ctx, keyName).Result()
		result.WriteString(fmt.Sprintf("List Length: %d\n", llen))
	case "set":
		scard, _ := client.SCard(ctx, keyName).Result()
		result.WriteString(fmt.Sprintf("Set Members: %d\n", scard))
	case "zset":
		zcard, _ := client.ZCard(ctx, keyName).Result()
		result.WriteString(fmt.Sprintf("Sorted Set Members: %d\n", zcard))
	}

	return mcp.NewToolResultText(result.String()), nil
}

func handleRedisRunCommand(connStr string, query string, confirm bool) (*mcp.CallToolResult, error) {
	if query == "" {
		return errResult("query is required")
	}

	parts := strings.Fields(query)
	if len(parts) == 0 {
		return mcp.NewToolResultText("Empty query"), nil
	}

	command := strings.ToUpper(parts[0])
	isWrite := command != "GET" && command != "MGET" && command != "HGET" && command != "HGETALL" &&
		command != "HMGET" && command != "HKEYS" && command != "HVALS" && command != "HLEN" &&
		command != "LRANGE" && command != "LLEN" && command != "LINDEX" && command != "SMEMBERS" &&
		command != "SCARD" && command != "SISMEMBER" && command != "ZRANGE" && command != "ZCARD" &&
		command != "ZSCORE" && command != "ZREVRANGE" && command != "TYPE" && command != "TTL" &&
		command != "EXISTS" && command != "SCAN" && command != "INFO" && command != "DBSIZE" && command != "PING"

	if isWrite && !confirm {
		return mcp.NewToolResultText(fmt.Sprintf("CONFIRMATION REQUIRED: You are about to execute a WRITE operation:\n\n%s\n\nSet confirm=true to proceed.", query)), nil
	}

	client, err := getRedisClient(connStr)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error connecting to redis: %v", err)), nil
	}

	var args []interface{}
	for _, p := range parts {
		args = append(args, p)
	}

	ctx := context.Background()
	res, err := client.Do(ctx, args...).Result()
	if err != nil {
		return mcp.NewToolResultText(fmt.Sprintf("Redis Error: %v", err)), nil
	}

	jsonBytes, _ := json.MarshalIndent(map[string]interface{}{"result": res}, "", "  ")
	return mcp.NewToolResultText(fmt.Sprintf("Result:\n```json\n%s\n```", string(jsonBytes))), nil
}