package kit

import (
	"log"
	"regexp"
	"strconv"
	"strings"
)

type Response struct {
	Code    int         `json:"code"`
	Data    interface{} `json:"data"`
	Message string      `json:"message"`
}

var reError = regexp.MustCompile(`\[(\d+)\]`)

const unknowErrCode = 1

func (r *Response) SetCode(errSrc error) {
	r.Code = unknowErrCode
	if errSrc == nil {
		r.Code = 0
		return
	}
	str := reError.FindAllStringSubmatch(errSrc.Error(), 1)
	if len(str) >= 1 && len(str[0]) >= 2 {
		code, err := strconv.Atoi(str[0][1])
		if err != nil {
			log.Println(err)
			return
		}
		r.Code = code
		r.Message = strings.Replace(errSrc.Error(), str[0][0], ``, 1)
		return
	}
	r.Message = errSrc.Error()
}
