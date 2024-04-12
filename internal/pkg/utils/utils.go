package utils

import (
	"math/rand"
	"os"
	"reflect"
	"runtime"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
	"unsafe"

	"github.com/google/uuid"
	"github.com/pkg/errors"
)

func init() {
	// 初始化伪随机数发生器
	rand.Seed(time.Now().UnixNano())
}

// UUID  return an uuid string
func UUID() string {
	return uuid.New().String()
}

// FileExists check if file exists at given path
func FileExists(path string) bool {
	fileInfo, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	if fileInfo.IsDir() {
		return false
	}
	return true
}

// CreateDirectory create directory if it not exists
func CreateDirectory(path string) error {
	fileInfo, err := os.Stat(path)
	if os.IsNotExist(err) {
		return os.MkdirAll(path, 0755)
	}
	if !fileInfo.IsDir() {
		return errors.New("file exists at directory path")
	}
	return nil
}

// DeleteDirectory remove directory if it exists
func DeleteDirectory(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil
	}
	return os.RemoveAll(path)
}

// GetFunctionName return name for func
func GetFunctionName(i interface{}) string {
	name := runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
	data := strings.Split(name, "/")
	return data[len(data)-1]
}

// 不包含o和l，避免与0和1混淆
const letters = "0123456789abcdefghijkmnpqrstuvwxyz"

// GetRandString return random string with given seed and length
func GetRandString(n int) string {
	a0 := time.Now().UnixNano()
	var data []byte
	for i := 0; i < n; i++ {
		index := int(rand.Int63()^a0) % len(letters)
		data = append(data, letters[index])
	}
	return string(data)
}

// FormatTime 格式化时间
func FormatTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

// CheckName 检查名称是否符合规格，要求：支持 1～127 位字符（1个汉字等于2个字符），必须以字母或中文开头，可以包含字母、数字、下划线（_）、中划线（-）、点(.)
func CheckName(n string) error {
	runeStr := []rune(n)
	if len(runeStr) == 0 {
		return errors.New("length mismatch")
	}
	ok := unicode.IsLetter(runeStr[0]) || unicode.Is(unicode.Han, runeStr[0])
	if !ok {
		return errors.New("format mismatch")
	}
	count := 0
	for _, r := range runeStr {
		if unicode.Is(unicode.Han, r) {
			count += 2
			continue
		}
		if unicode.IsLetter(r) || unicode.IsNumber(r) || r == '_' || r == '-' || r == '.' {
			count += 1
			continue
		}
		return errors.New("format mismatch")
	}
	if count > 127 {
		return errors.New("length mismatch")
	}
	return nil
}

type passwdRuneType uint8

const (
	passwdRuneTypeAlpha passwdRuneType = iota
	passwdRuneTypeNum
	passwdRuneTypeSymbol
)

var passwordRuneMap = map[rune]passwdRuneType{}

func init() {
	a := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := "0123456789"
	c := "!@%^-_=+[{}]:,./?"
	for _, v := range a {
		passwordRuneMap[v] = 0
	}
	for _, v := range b {
		passwordRuneMap[v] = 1
	}
	for _, v := range c {
		passwordRuneMap[v] = 2
	}
}

// CheckPassword 检查密码是否符合规则，要求：支持8~32字符，至少包含数字、字母及特殊字符中的2种
func CheckPassword(n string) error {
	count := utf8.RuneCountInString(n)
	if count < 8 || count > 32 {
		return errors.New("length mismatch")
	}
	alpha := 0
	num := 0
	symbol := 0
	for _, v := range n {
		vType, ok := passwordRuneMap[v]
		if !ok {
			return errors.New("format mismatch")
		}
		switch vType {
		case passwdRuneTypeAlpha:
			alpha = 1
		case passwdRuneTypeNum:
			num = 1
		case passwdRuneTypeSymbol:
			symbol = 1
		}
	}
	v := alpha + num + symbol
	if v < 2 {
		return errors.New("format mismatch")
	}
	return nil
}

func GetStructPtrUnExportedField(source interface{}, fieldName string) reflect.Value {
	// 获取非导出字段反射对象
	v := reflect.ValueOf(source).Elem().FieldByName(fieldName)
	// 构建指向该字段的可寻址（addressable）反射对象
	return reflect.NewAt(v.Type(), unsafe.Pointer(v.UnsafeAddr())).Elem()
}

func SetStructPtrUnExportedStrField(source interface{}, fieldName string, fieldVal interface{}) (err error) {
	v := GetStructPtrUnExportedField(source, fieldName)
	rv := reflect.ValueOf(fieldVal)
	if v.Kind() != rv.Kind() {
		return errors.Errorf("invalid kind: expected kind %v: got kind: %v", v.Kind(), rv.Kind())
	}
	// 修改非导出字段值
	v.Set(rv)
	return nil
}

// value转指针函数，用于k8s一些value参数

func GetInt32Ptr(i int32) *int32 {
	return &i
}

func GetUint32Ptr(i uint32) *uint32 {
	return &i
}

func GetInt64Ptr(i int64) *int64 {
	return &i
}

func GetUint64Ptr(i uint64) *uint64 {
	return &i
}

func GetBoolPtr(i bool) *bool {
	return &i
}
