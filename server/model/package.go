package model

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"database/sql/driver"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/BernardSimon/etl-go/server/config"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm/schema"

	"go.uber.org/zap"
	"gorm.io/gorm/logger"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var DB *gorm.DB

func MigrateDb() error {
	err := DB.AutoMigrate(&DataSource{}, &Variable{}, &Task{}, &TaskRecord{}, &File{}, &TaskRecordFile{})
	if err != nil {
		return err
	}
	return nil
}

func InitDb() error {
	// 初始化 SQLite 数据库连接
	dB, err := gorm.Open(sqlite.Open("./data.db"), &gorm.Config{
		Logger:      &sqlLogger{},
		PrepareStmt: true, // 开启预编译语句缓存
	})
	if err != nil {
		return fmt.Errorf("failed to connect database: %w", err)
	}

	// 获取底层sql.DB对象配置连接池
	sqlDB, err := dB.DB()
	if err != nil {
		return fmt.Errorf("failed to obtain database instance: %w", err)
	}

	// 测试数据库连通性
	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("database connection test failed: %w", err)
	}
	DB = dB
	return nil
}

type Model struct {
	ID        string      `gorm:"size:36;primarykey" json:"id"`
	UpdatedAt *CustomTime `json:"updated_at"`
	CreatedAt *CustomTime `json:"created_at"`
}

func (m *Model) BeforeCreate(_ *gorm.DB) (err error) {
	if m.ID == "" {
		m.ID = uuid.New().String()
	}
	return
}

// sqlLogger 自定义 zap 日志记录器
type sqlLogger struct {
	logger.Interface
}

// LogMode 实现 logger.Interface 接口
func (l *sqlLogger) LogMode(_ logger.LogLevel) logger.Interface {
	return l
}

// Info 实现 logger.Interface 接口
func (l *sqlLogger) Info(_ context.Context, msg string, args ...interface{}) {
	zap.L().Info(msg, zap.Any("content", args), zap.String("service", "sql"))
}

// Warn 实现 logger.Interface 接口
func (l *sqlLogger) Warn(_ context.Context, msg string, args ...interface{}) {
	zap.L().Warn(msg, zap.Any("content", args), zap.String("service", "sql"))
}

// Error 实现 logger.Interface 接口
func (l *sqlLogger) Error(_ context.Context, msg string, args ...interface{}) {
	zap.L().Error(msg, zap.Any("content", args), zap.String("service", "sql"))
}

// Trace 实现 logger.Interface 接口
func (l *sqlLogger) Trace(_ context.Context, begin time.Time, fc func() (string, int64), err error) {
	sql, rows := fc()
	elapsed := time.Since(begin)
	fields := []zap.Field{
		zap.Any("content", []zap.Field{zap.Int64("rows", rows), zap.String("sql", sql), zap.Duration("elapsed", elapsed)}),
		zap.String("service", "sql"),
	}
	if err != nil {
		fields = append(fields, zap.Error(err))
		zap.L().Error("sql trace error", fields...)
	}
	if elapsed > time.Millisecond*500 {
		zap.L().Warn("sql trace slow", fields...)
	}
}

// EncryptionSerializer 加解密序列化器
type EncryptionSerializer struct {
}

// Scan 实现 Scan 方法
func (e EncryptionSerializer) Scan(ctx context.Context, field *schema.Field, dst reflect.Value, dbValue interface{}) (err error) {
	if dbValue != nil {
		var cipherText []byte
		switch v := dbValue.(type) {
		case string:
			cipherText = []byte(v)
		case []byte:
			cipherText = v
		default:
			return fmt.Errorf("failed to read cipher text value: %#v", dbValue)
		}

		key, err := hex.DecodeString(config.Config.AesKey)
		if err != nil {
			return fmt.Errorf("failed to decode key: %w", err)
		}
		decodedCipherText, err := base64.StdEncoding.DecodeString(string(cipherText))
		if err != nil {
			return fmt.Errorf("failed to decode cipher text: %w", err)
		}
		block, err := aes.NewCipher(key)
		if err != nil {
			return fmt.Errorf("failed to create cipher: %w", err)
		}
		if len(decodedCipherText) < aes.BlockSize {
			return fmt.Errorf("cipher text too short")
		}
		iv := decodedCipherText[:aes.BlockSize]
		decodedCipherText = decodedCipherText[aes.BlockSize:]
		mode := cipher.NewCBCDecrypter(block, iv)
		mode.CryptBlocks(decodedCipherText, decodedCipherText)
		// 去除 PKCS5 填充
		padding := decodedCipherText[len(decodedCipherText)-1]
		decodedCipherText = decodedCipherText[:len(decodedCipherText)-int(padding)]

		// 判断字段类型是否是结构体或指针
		fieldValue := field.ReflectValueOf(ctx, dst)
		if fieldValue.Kind() == reflect.Ptr && !fieldValue.IsNil() {
			fieldValue = fieldValue.Elem()
		}

		// 如果是结构体类型则进行 JSON 解析
		if fieldValue.Kind() == reflect.Struct || (fieldValue.Kind() == reflect.Ptr && fieldValue.Type().Elem().Kind() == reflect.Struct) {
			err = json.Unmarshal(decodedCipherText, fieldValue.Addr().Interface())
			if err != nil {
				return fmt.Errorf("failed to unmarshal struct: %w", err)
			}
		} else {
			// 根据字段的具体类型设置值
			switch fieldValue.Kind() {
			case reflect.String:
				fieldValue.SetString(string(decodedCipherText))
			case reflect.Slice:
				if fieldValue.Type().Elem().Kind() == reflect.Uint8 {
					// []byte 类型
					fieldValue.SetBytes(decodedCipherText)
				} else {
					// 其他切片类型（如 _type.KeyValues）通过 JSON 反序列化处理
					err = json.Unmarshal(decodedCipherText, fieldValue.Addr().Interface())
					if err != nil {
						return fmt.Errorf("failed to unmarshal slice: %w", err)
					}
				}
			default:
				return fmt.Errorf("unsupported field type: %v", fieldValue.Kind())
			}
		}
	}
	return
}

// Value 实现 Value 方法
func (e EncryptionSerializer) Value(_ context.Context, _ *schema.Field, _ reflect.Value, fieldValue interface{}) (interface{}, error) {
	var plainText string
	if fieldValue != nil {
		switch v := fieldValue.(type) {
		case string:
			plainText = v
		case []byte:
			plainText = string(v)
		default:
			b, err := json.Marshal(v)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal value: %w", err)
			}
			plainText = string(b)
		}

		// 加密逻辑
		key, _ := hex.DecodeString(config.Config.AesKey)
		block, err := aes.NewCipher(key)
		if err != nil {
			return nil, fmt.Errorf("failed to create cipher: %w", err)
		}

		// PKCS5 填充
		padding := aes.BlockSize - len(plainText)%aes.BlockSize
		padText := bytes.Repeat([]byte{byte(padding)}, padding)
		plainTextBytes := append([]byte(plainText), padText...)

		iv := make([]byte, aes.BlockSize)
		if _, err := rand.Read(iv); err != nil {
			return nil, fmt.Errorf("failed to generate IV: %w", err)
		}

		mode := cipher.NewCBCEncrypter(block, iv)
		cipherText := make([]byte, len(plainTextBytes))
		mode.CryptBlocks(cipherText, plainTextBytes)

		// 将 IV 和密文拼接在一起
		cipherText = append(iv, cipherText...)

		return base64.StdEncoding.EncodeToString(cipherText), nil
	}
	return nil, nil
}

type CustomTime struct {
	time.Time
}

// UnmarshalJSON 解析 JSON 数据到 CustomTime
func (ct *CustomTime) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)
	if s == "" {
		return nil
	}

	// 尝试解析为 Unix 时间戳（毫秒）
	if ts, err := strconv.ParseInt(s, 10, 64); err == nil {
		t := time.Unix(ts/1000, 0)
		ct.Time = t.Local()
		return nil
	}

	// 尝试解析为日期时间格式 "2006-01-02 15:04:05"
	t, err := time.ParseInLocation("2006-01-02 15:04:05", s, time.Local)
	if err == nil {
		ct.Time = t
		return nil
	}

	// 尝试解析为日期格式 "2006-01-02"
	t, err = time.ParseInLocation("2006-01-02", s, time.Local)
	if err == nil {
		ct.Time = t
		return nil
	}

	return fmt.Errorf("unsupported time format: %s", s)
}

func (ct *CustomTime) MarshalJSON() ([]byte, error) {
	if ct == nil {
		return []byte(nil), nil
	}
	if ct.Time.IsZero() {
		return []byte("0000-00-00 00:00:00"), nil
	}
	return []byte(fmt.Sprintf(`"%s"`, ct.Time.Format("2006-01-02 15:04:05"))), nil
}

// Value 实现 Valuer 接口，将 CustomTime 转换为 time.Time
func (ct *CustomTime) Value() (driver.Value, error) {
	if ct == nil || ct.Time.IsZero() {
		return nil, nil
	}
	return ct.Time, nil
}

// Scan 实现 Scanner 接口，将数据库中的时间值转换为 CustomTime
func (ct *CustomTime) Scan(value interface{}) error {
	if value == nil {
		ct.Time = time.Time{}
		return nil
	}

	switch v := value.(type) {
	case time.Time:
		ct.Time = v
		return nil
	case []byte:
		t, err := time.Parse("2006-01-02 15:04:05", string(v))
		if err != nil {
			return err
		}
		ct.Time = t
		return nil
	default:
		return fmt.Errorf("unsupported type: %T", value)
	}
}
