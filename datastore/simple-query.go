package datastore

import "time"

func NewQuery() *SimpleQuery {
	return &SimpleQuery{
		Conditions: &SimpleQueryConditionGroup{
			Operator: "and",
		},
	}
}

// SimpleQuery is a simple way to describe conditions and a query
type SimpleQuery struct {
	Size          int
	Offset        int
	Colums        []string
	Conditions    *SimpleQueryConditionGroup
	SortBy        []*SortBy
	RecurseConfig *SimpleQueryRecurse
}

type SimpleQueryRecurse struct {
	FromField string
	ToField   string
}

// Recurse will setup this query as a recursive query. The FromField is the field that
// that is the starting point. The ToField is the field that is the ending point.
// For example, to go UP a hierarchy, the FromField would be the parent field and the
// to the child field, e.g. `Recurse("parent", "id")`. To go DOWN the hierarchy, the
// FromField would be the child field and the ToField would be the parent field, e.g.
// `Recurse("id", "parent")`
func (sq *SimpleQuery) Recurse(FromField string, ToField string) *SimpleQuery {
	sq.RecurseConfig = &SimpleQueryRecurse{
		FromField: FromField,
		ToField:   ToField,
	}
	return sq
}

type SimpleQueryCondition struct {
	Type    string
	Data    []string
	DataMap map[string]interface{}
}

type SimpleQueryConditionGroup struct {
	Operator   string //and, or, no
	Conditions []*SimpleQueryCondition
	Groups     []*SimpleQueryConditionGroup
}

type SortBy struct {
	Field      string
	Descending bool
}

func (sqc *SimpleQueryCondition) Set(key string, value interface{}) {
	if sqc.DataMap == nil {
		sqc.DataMap = make(map[string]interface{})
	}
	sqc.DataMap[key] = value
}

func (sqc *SimpleQueryCondition) GetString(key string) string {
	if sqc.DataMap != nil {
		val := sqc.DataMap[key]
		return val.(string)
	}
	return ""
}

func (sqc *SimpleQueryCondition) GetStringArr(key string) []string {
	if sqc.DataMap != nil {
		val := sqc.DataMap[key]
		return val.([]string)
	}
	return nil
}

func (sqc *SimpleQueryCondition) GetInt(key string) int {
	if sqc.DataMap != nil {
		val := sqc.DataMap[key]
		return val.(int)
	}
	return 0
}

func (sqc *SimpleQueryCondition) GetFloat(key string) float64 {
	if sqc.DataMap != nil {
		val := sqc.DataMap[key]
		return val.(float64)
	}
	return 0
}

func (sqc *SimpleQueryCondition) GetDate(key string) (val time.Time) {
	if sqc.DataMap != nil {
		val = sqc.DataMap[key].(time.Time)
	}
	return
}

func (cg *SimpleQueryConditionGroup) Null(field string) {
	cg.Conditions = append(cg.Conditions, &SimpleQueryCondition{
		Type: "null",
		Data: []string{field},
	})
}

func (cg *SimpleQueryConditionGroup) Contains(field string, value string) {
	cg.Conditions = append(cg.Conditions, &SimpleQueryCondition{
		Type: "contains",
		Data: []string{field, value},
	})
}

func (cg *SimpleQueryConditionGroup) In(field string, value string) {
	c := &SimpleQueryCondition{Type: "in"}
	c.Data = []string{field, value}
	c.Set("value", value)
	c.Set("field", field)
	cg.Conditions = append(cg.Conditions, c)
}

func (cg *SimpleQueryConditionGroup) AnyIn(field string, values []string) {
	c := &SimpleQueryCondition{Type: "anyin"}
	c.Data = []string{field}
	c.Set("value", values)
	c.Set("field", field)
	cg.Conditions = append(cg.Conditions, c)
}

func (cg *SimpleQueryConditionGroup) Includes(field string, values []string) {
	c := &SimpleQueryCondition{Type: "includes"}
	c.Data = []string{field}
	c.Set("value", values)
	c.Set("field", field)
	cg.Conditions = append(cg.Conditions, c)
}

func (cg *SimpleQueryConditionGroup) After(field string, value time.Time) {
	c := &SimpleQueryCondition{Type: "after"}
	c.Data = []string{field}
	c.Set("field", field)
	c.Set("value", value)
	cg.Conditions = append(cg.Conditions, c)
}

func (cg *SimpleQueryConditionGroup) Before(field string, value time.Time) {
	c := &SimpleQueryCondition{Type: "before"}
	c.Data = []string{field}
	c.Set("field", field)
	c.Set("value", value)
	cg.Conditions = append(cg.Conditions, c)
}

func (cg *SimpleQueryConditionGroup) Equals(field string, value string) {
	cg.Conditions = append(cg.Conditions, &SimpleQueryCondition{
		Type: "eq",
		Data: []string{field, value},
	})
}
func (cg *SimpleQueryConditionGroup) Between(field string, gte string, lte string) {
	cg.Conditions = append(cg.Conditions, &SimpleQueryCondition{
		Type: "between",
		Data: []string{field, gte, lte},
	})
}
func (cg *SimpleQueryConditionGroup) LessThan(field string, value string) {
	cg.Conditions = append(cg.Conditions, &SimpleQueryCondition{
		Type: "lt",
		Data: []string{field, value},
	})
}
func (cg *SimpleQueryConditionGroup) LessThanOrEqual(field string, value string) {
	cg.Conditions = append(cg.Conditions, &SimpleQueryCondition{
		Type: "lte",
		Data: []string{field, value},
	})
}
func (cg *SimpleQueryConditionGroup) GreaterThan(field string, value string) {
	cg.Conditions = append(cg.Conditions, &SimpleQueryCondition{
		Type: "gt",
		Data: []string{field, value},
	})
}
func (cg *SimpleQueryConditionGroup) GreaterThanOrEqual(field string, value string) {
	cg.Conditions = append(cg.Conditions, &SimpleQueryCondition{
		Type: "gte",
		Data: []string{field, value},
	})
}
func (cg *SimpleQueryConditionGroup) Exists(field string, value string) {
	cg.Conditions = append(cg.Conditions, &SimpleQueryCondition{
		Type: "?",
		Data: []string{field, value},
	})
}
func (cg *SimpleQueryConditionGroup) And() *SimpleQueryConditionGroup {
	grp := &SimpleQueryConditionGroup{
		Operator: "and",
	}
	cg.Groups = append(cg.Groups, grp)
	return grp
}
func (cg *SimpleQueryConditionGroup) Or() *SimpleQueryConditionGroup {
	grp := &SimpleQueryConditionGroup{
		Operator: "or",
	}
	cg.Groups = append(cg.Groups, grp)
	return grp
}
func (cg *SimpleQueryConditionGroup) Not() *SimpleQueryConditionGroup {
	grp := &SimpleQueryConditionGroup{
		Operator: "not",
	}
	cg.Groups = append(cg.Groups, grp)
	return grp
}
func (cg *SimpleQueryConditionGroup) IsAny(field string, values []string) *SimpleQueryConditionGroup {
	grp := &SimpleQueryConditionGroup{
		Operator: "or",
	}
	for _, v := range values {
		grp.Equals(field, v)
	}
	cg.Groups = append(cg.Groups, grp)
	return grp
}
