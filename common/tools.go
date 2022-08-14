package common

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"encoding/binary"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"github.com/fatih/structs"
	"io"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/satori/go.uuid"
)

type Tools struct {
}

var (
	t     *Tools
	once  sync.Once
	tools *Tools
)

func init() {
	tools = NewTools()
}

/**
 * 返回单例实例
 * @method New
 */
func NewTools() *Tools {
	once.Do(func() { //只执行一次
		t = &Tools{}
	})

	return t
}

/**
 * md5 加密
 * @method MD5
 * @param  {[type]} data string [description]
 */
func (t *Tools) MD5(data string) string {
	m := md5.New()
	io.WriteString(m, data)

	return fmt.Sprintf("%x", m.Sum(nil))
}

/**
 * string转换int
 * @method parseInt
 * @param  {[type]} b string        [description]
 * @return {[type]}   [description]
 */
func (t *Tools) ParseInt64(b string, defInt int64) int64 {
	id, err := strconv.ParseInt(b, 10, 64)
	if err != nil {
		return defInt
	} else {
		return id
	}
}

/**
 * 结构体转换成map对象
 * @method func
 * @param  {[type]} t *Tools        [description]
 * @return {[type]}   [description]
 */
func (t *Tools) GetDateNowString() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

/**
 * 结构体转换成map对象
 * @method func
 * @param  {[type]} u *Utils        [description]
 * @return {[type]}   [description]
 */
func (t *Tools) StructToMap(obj interface{}) map[string]interface{} {
	return structs.Map(obj)
}

func (t *Tools) CopyStruct(dst, src interface{}) error {
	jsonStr, err := json.Marshal(dst)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(jsonStr, src); err != nil {
		return err
	}

	return nil

	//dstValue := reflect.ValueOf(dst)
	//if dstValue.Kind() != reflect.Ptr {
	//	err := errors.New("dst isn't a pointer to struct")
	//	return err
	//}
	//
	//dstElem := dstValue.Elem()
	//if dstElem.Kind() != reflect.Struct {
	//	err := errors.New("pointer doesn't point to struct")
	//	return err
	//}
	//
	//srcValue := reflect.ValueOf(src)
	//srcType := reflect.TypeOf(src)
	//if srcType.Kind() != reflect.Struct {
	//	err := errors.New("src isn't struct")
	//	return err
	//}
	//
	//for i := 0; i < srcType.NumField(); i++ {
	//	sf := srcType.Field(i)
	//	sv := srcValue.FieldByName(sf.Name)
	//	if dv := dstElem.FieldByName(sf.Name); dv.IsValid() && dv.CanSet() {
	//		dv.Set(sv)
	//	}
	//}
	//
	//return nil
}

/**
 * 判断手机号码
 * @method func
 * @param  {[type]} u *Utils        [description]
 * @return {[type]}   [description]
 */
func (t *Tools) IsMobile(mobile string) bool {

	reg := `^1([38][0-9]|14[57]|5[^4])\d{8}$`

	rgx := regexp.MustCompile(reg)

	return rgx.MatchString(mobile)
}

/**
 * 验证密码
 * @method func
 * @param  {[type]} u *Utils        [description]
 * @return {[type]}   [description]
 */
func (t *Tools) CheckPassword(password, metaPassword string) bool {

	return strings.EqualFold(password, metaPassword)
}

/**
 * 生成随机字符串
 * @method func
 * @param  {[type]} u *Utils        [description]
 * @return {[type]}   [description]
 */
func (t *Tools) GetRandomString(n int) string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Int63()%int64(len(letterBytes))]
	}

	return string(b)
}

/**
 * 生成用户Redis key
 * @method func
 * @param  {[type]} u *Utils        [description]
 * @return {[type]}   [description]
 */
func (t *Tools) UserRedisKey(userId int64) string {
	userKey := fmt.Sprintf("user_login_%d", userId)

	return userKey
}

/**
 * 生成用户Token
 * @method func
 * @param  {[type]} u *Utils        [description]
 * @return {[type]}   [description]
 */
func (t *Tools) GenerateToken(n int) (string, error) {
	token, err := t.GenerateRandomString(n)
	return token, err
}

func (t *Tools) GenerateRandomString(s int) (string, error) {
	b, err := t.GenerateRandomBytes(s)
	return base64.URLEncoding.EncodeToString(b), err
}

func (t *Tools) GenerateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func (t *Tools) GenerateFileName(path, fileName string) string {
	now := time.Now().UnixNano()
	random := t.MD5(t.GetRandomString(12))

	number := len(random)

	path = fmt.Sprintf("%s/%s/%s",
		strings.Trim(path, "/"),
		string([]byte(random)[:6]),
		string([]byte(random)[number-6:number]))

	t.CreatedDir(path, os.ModePerm)

	return fmt.Sprintf("%s/%d_%s%s",
		path,
		now,
		string([]byte(random)[10:20]),
		filepath.Ext(fileName))
}

func (t *Tools) CreatedDir(dir string, mode os.FileMode) {
	ok, err := t.PathExists(dir)
	if err == nil && !ok {
		os.MkdirAll(dir, mode)
	}
}

func (t *Tools) SnakeString(s string) string {
	data := make([]byte, 0, len(s)*2)
	j := false
	num := len(s)

	for i := 0; i < num; i++ {
		d := s[i]
		if i > 0 && d >= 'A' && d <= 'Z' && j {
			data = append(data, '_')
		}
		if d != '_' {
			j = true
		}

		data = append(data, d)
	}

	return strings.ToLower(string(data[:]))
}

func (t *Tools) GetNowMillisecond() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func (t *Tools) GetAddMillisecond(add time.Duration) int64 {
	return time.Now().Add(add).UnixNano() / int64(time.Millisecond)
}

func (t *Tools) GetAcceptLanguage(acceptLanguage string) string {
	language := "zh-CN"

	lang := strings.Split(acceptLanguage, ";")
	if len(lang) >= 1 {
		langs := strings.Split(lang[0], ",")
		language = langs[0]
	}

	return language
}

func (t *Tools) PathExists(path string) (bool, error) {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		} else {
			return false, err
		}
	} else {
		return true, nil
	}
}

func (t *Tools) GenerateRangeNum(min, max int) int {
	rand.Seed(time.Now().Unix())
	randNum := rand.Intn(max - min)
	randNum = randNum + min

	return randNum
}

func (t *Tools) CreateUUID() string {
	return uuid.NewV4().String()
}

func (t *Tools) SetOffset(page, size int64) int64 {
	offset := (page - 1) * size

	return offset
}

func (t *Tools) Merge(data []string, dist ...[]string) []string {
	for _, list := range dist {
		for _, v := range list {
			data = append(data, v)
		}
	}

	return data
}

func (t *Tools) GetStatus(err error) string {
	if err != nil {
		return "errors"
	}

	return "success"
}

func (t *Tools) LoopCall(structs ...interface{}) {
	for _, v := range structs {
		classType := reflect.TypeOf(v)
		classValue := reflect.ValueOf(v)

		for i := 0; i < classType.NumMethod(); i++ {
			m := classValue.MethodByName(classType.Method(i).Name)
			if m.IsValid() {
				var params []reflect.Value
				m.Call(params)
			}
		}
	}
}

func (t *Tools) UniqueInt(elements []int) []int {
	encountered := map[int]bool{}
	var result []int

	for v := range elements {
		if !encountered[elements[v]] == true {
			encountered[elements[v]] = true
			result = append(result, elements[v])
		}
	}

	return result
}

func (t *Tools) UniqueString(elements []string) []string {
	encountered := map[string]bool{}
	var result []string

	for v := range elements {
		if !encountered[elements[v]] == true {
			encountered[elements[v]] = true
			result = append(result, elements[v])
		}
	}

	return result
}

func (t *Tools) Int64ToBytes(i int64) []byte {
	var buf = make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(i))

	return buf
}

func (t *Tools) Uint64ToBytes(i uint64) []byte {
	var buf = make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(i))

	return buf
}

func (t *Tools) StringToBytes(str string) []byte {
	return []byte(str)
}

func (t *Tools) ReverseString(s string) string {
	runes := []rune(s)
	for from, to := 0, len(runes)-1; from < to; from, to = from+1, to-1 {
		runes[from], runes[to] = runes[to], runes[from]
	}

	return string(runes)
}

func (t *Tools) Ucfirst(str string) string {
	for i, v := range str {
		return string(unicode.ToUpper(v)) + str[i+1:]
	}

	return ""
}

func (t *Tools) UintToString(data uint64) string {
	return strconv.FormatInt(int64(data), 10)
}

func (t *Tools) InterfaceToBytes(body interface{}) ([]byte, error) {
	var buf bytes.Buffer

	enc := gob.NewEncoder(&buf)
	err := enc.Encode(body)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (t *Tools) Implode(list interface{}, seq string) string {
	listValue := reflect.Indirect(reflect.ValueOf(list))
	if listValue.Kind() != reflect.Slice {
		return ""
	}

	count := listValue.Len()
	listStr := make([]string, 0, count)
	for i := 0; i < count; i++ {
		v := listValue.Index(i)

		if str, err := t.getValue(v); err == nil {
			listStr = append(listStr, str)
		}
	}

	return strings.Join(listStr, seq)
}

func (t *Tools) getValue(value reflect.Value) (res string, err error) {
	switch value.Kind() {
	case reflect.Ptr:
		res, err = t.getValue(value.Elem())
	default:
		res = fmt.Sprint(value.Interface())
	}

	return
}

func (t *Tools) CamelString(s string) string {
	data := make([]byte, 0, len(s))

	j := false
	k := false
	num := len(s) - 1

	for i := 0; i <= num; i++ {
		d := s[i]
		if k == false && d >= 'A' && d <= 'Z' {
			k = true
		}

		if d >= 'a' && d <= 'z' && (j || k == false) {
			d = d - 32
			j = false
			k = true
		}

		if k && d == '_' && num > i && s[i+1] >= 'a' && s[i+1] <= 'z' {
			j = true
			continue
		}

		data = append(data, d)
	}

	return string(data[:])
}

func (t *Tools) Contains(array interface{}, val interface{}) bool {
	switch reflect.TypeOf(array).Kind() {
	case reflect.Slice:
		{
			s := reflect.ValueOf(array)
			for i := 0; i < s.Len(); i++ {
				if reflect.DeepEqual(val, s.Index(i).Interface()) {
					return true
				}
			}
		}
	}

	return false
}

func (t *Tools) ExplodeUint64(step string, str string) []uint64 {
	var result = []uint64{}

	if str == "" {
		return result
	}
	strArray := strings.Split(str, step)

	for _, s := range strArray {
		i, _ := strconv.Atoi(s)
		result = append(result, uint64(i))
	}

	return result
}

func (t *Tools) ExplodeString(step, str string) []string {
	strArray := strings.Split(str, step)

	var result []string
	for _, str := range strArray {
		result = append(result, str)
	}

	return result
}

func (t *Tools) IsNil(v interface{}) bool {
	return v == nil || (reflect.ValueOf(v).Kind() == reflect.Ptr && reflect.ValueOf(v).IsNil())
}

func (t *Tools) SubString(str string, begin, length int, suffix ...string) string {
	ext := ""
	if len(suffix) == 1 {
		ext = suffix[0]
	}

	rs := []rune(str)
	lth := len(rs)

	if begin < 0 {
		begin = 0
	}
	if begin >= lth {
		begin = lth
	}
	end := begin + length

	if end > lth {
		end = lth
	}

	if lth <= length {
		ext = ""
	}

	return string(rs[begin:end]) + ext
}

func (t *Tools) UnsetUin64(nums []uint64, val uint64) []uint64 {
	if len(nums) == 0 {
		return nums
	}

	index := 0
	for index < len(nums) {
		if nums[index] == val {
			nums = append(nums[:index], nums[index+1:]...)
			continue
		}
		index++
	}
	return nums
}

func (t *Tools) Do(attempts int) time.Duration {
	if attempts > 13 {
		return 2 * time.Minute
	}

	return time.Duration(math.Pow(float64(attempts), math.E)) * time.Millisecond * 100
}

func (t *Tools) TimeNowUTC() time.Time {
	return time.Now().UTC().Round(time.Millisecond)
}
