package mysql

import (
	"fmt"
	"github.com/yousinn/go-common/common"
	"gorm.io/gorm"
	"reflect"
)

type (
	Service struct {
		field string
	}

	QueryResult struct {
		Total  int64       `json:"total"`
		Size   int         `json:"size"`
		List   interface{} `json:"list"`
		LastId uint64      `json:"lastId"`
	}

	QueryPage struct {
		Page   int64  `json:"page" form:"page" binding:"omitempty,numeric"`
		Size   int    `json:"size" form:"size"  binding:"omitempty,numeric"`
		LastId uint64 `json:"lastId" form:"lastId" binding:"omitempty,numeric"`
	}
)

func NewMysqlService(lastIdField ...string) *Service {
	return &Service{
		field: getLastIdFieldAlias(lastIdField...),
	}
}

func (s *Service) QueryResList(
	query *gorm.DB,
	tableName string, list interface{},
	lastId uint64, size int) (*QueryResult, error) {

	var total int64

	queryCount := query
	queryCount.Table(tableName).Count(&total)
	if total > 0 {
		if lastId > 0 {
			query = query.Where(fmt.Sprintf("%s < ?", s.field), lastId)
		}

		err = query.Table(tableName).Limit(s.GetSize(size)).Find(list).Error
		if err != nil {
			return nil, err
		}
	}

	result := &QueryResult{
		List:   list,
		Total:  total,
		Size:   s.GetSize(size),
		LastId: s.getLastId(list, total, s.GetSize(size)),
	}

	return result, err
}

const (
	defaultSize = 20
)

func (s *Service) PageQueryReflect(
	query *gorm.DB,
	tableName string, list interface{},
	lastId uint64, size int,
	class interface{},
	name string) (*QueryResult, error) {

	var total int64
	queryCount := query
	queryCount.Table(tableName).Count(&total)
	if total > 0 {
		if lastId > 0 {
			query = query.Where(fmt.Sprintf("%s < ?", s.field), lastId)
		}

		err := query.Table(tableName).Limit(s.GetSize(size)).Find(list).Error
		if err != nil {
			return nil, err
		}
	}

	result := &QueryResult{
		List:   list,
		Total:  total,
		Size:   s.GetSize(size),
		LastId: s.getLastId(list, total, s.GetSize(size)),
	}

	if class != nil && name != "" {
		r, err := s.Calls(class, name, list)
		if err != nil {
			return nil, err
		}

		result.List = r[0].Interface()
	}

	return result, err
}

func (s *Service) PageResult(
	list interface{},
	total int64, size int) (*QueryResult, error) {

	return &QueryResult{
		List:   list,
		Total:  total,
		Size:   s.GetSize(size),
		LastId: s.getLastId(list, total, s.GetSize(size)),
	}, nil
}

func (s *Service) getLastId(list interface{}, total int64, size int) uint64 {
	if total > int64(size) {
		return s.reflectLastId(list)
	}

	return 0
}

func (s *Service) reflectLastId(list interface{}) uint64 {
	items := reflect.ValueOf(list)
	if items.Kind() == reflect.Ptr {
		items := reflect.ValueOf(list).Elem()
		if items.Kind() == reflect.Slice {
			index := items.Len()
			if index == 0 {
				return 0
			}

			itemValue := items.Index(index - 1)

			var item reflect.Value
			switch itemValue.Kind() {
			case reflect.Struct:
				item = itemValue
			case reflect.Ptr:
				item = itemValue.Elem()
			}

			v := reflect.Indirect(item)
			fmt.Println("field", common.NewTools().CamelString(s.field))
			return v.FieldByName(common.NewTools().CamelString(s.field)).Uint()
		}
	}

	return 0
}

func getLastIdFieldAlias(field ...string) string {
	if len(field) > 0 {
		return field[0]
	}

	return "id"
}

func (s *Service) Calls(
	myClass interface{},
	name string,
	params ...interface{}) ([]reflect.Value, error) {

	myClassValue := reflect.ValueOf(myClass)
	m := myClassValue.MethodByName(name)

	if !m.IsValid() {
		err := fmt.Errorf("method not found param name: %s", name)
		return nil, err
	}

	in := make([]reflect.Value, len(params))
	for i, param := range params {
		in[i] = reflect.ValueOf(param)
	}

	return m.Call(in), nil
}

func (s *Service) GetSize(size int) int {
	if size <= 0 {
		size = defaultSize
	}

	return size
}
