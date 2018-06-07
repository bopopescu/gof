package core

import (
	"database/sql/driver"
	"fmt"
	"time"
)

//JSONTime ...
type JSONTime time.Time

// Value insert timestamp into mysql need this function.
func (p JSONTime) Value() (driver.Value, error) {
	var zeroTime time.Time
	var ti = time.Time(p)
	if ti.UnixNano() == zeroTime.UnixNano() {
		return nil, nil
	}
	return ti, nil
}

// Scan valueof time.Time
func (p *JSONTime) Scan(v interface{}) error {
	value, ok := v.(time.Time)
	if ok {
		*p = JSONTime(value)
		return nil
	}
	return fmt.Errorf("can not convert %v to timestamp", v)
}

// GobEncode implements the gob.GobEncoder interface.
func (p JSONTime) GobEncode() ([]byte, error) {
	return time.Time(p).MarshalBinary()
}

// GobDecode implements the gob.GobDecoder interface.
func (p *JSONTime) GobDecode(data []byte) error {
	s := time.Time(*p)
	p1 := &s
	return p1.UnmarshalBinary(data)
}

//MarshalJSON ...
func (p JSONTime) MarshalJSON() ([]byte, error) {
	if time.Time(p).IsZero() {
		return []byte(`""`), nil
	}
	data := make([]byte, 0)
	data = append(data, '"')
	data = time.Time(p).AppendFormat(data, "2006-01-02 15:04:05")
	data = append(data, '"')
	return data, nil
}

//UnmarshalJSON ...
func (p *JSONTime) UnmarshalJSON(data []byte) error {
	local, _ := time.ParseInLocation(`"`+"2006-01-02 15:04:05"+`"`, string(data), time.Local)
	*p = JSONTime(local)
	return nil
}

//String ...
func (p JSONTime) String() string {
	return time.Time(p).Format("2006-01-02 15:04:05")
}

//Todatetime ...
func Todatetime(in string) (JSONTime, error) {
	out, err := time.ParseInLocation("2006-01-02 15:04:05", in, time.Local)
	return JSONTime(out), err
}
