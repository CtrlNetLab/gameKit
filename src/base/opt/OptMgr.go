package opt

import (
	optcommon "gamekit/pkg/opt/scommon"
	"gamekit/src/base/logger"
	"gamekit/src/base/utils"
	"fmt"
	"reflect"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

/**
* @author: Tavin
* @date: 2024/4/24
 */

// APIResult 运营平台接口返回
type APIResult[T interface{}] struct {
	Code      optcommon.ResultCode
	Msg       string
	IsSuccess bool
	Data      T
}

// DoOperationRequest 运营平台接口请求
func DoOperationRequest[T proto.Message](optAddr string, apiId optcommon.APIId, reqData proto.Message) *APIResult[T] {

	var t T
	localApiResult := &APIResult[T]{
		IsSuccess: false,
	}
	if len(optAddr) == 0 {
		//请求地址不能为空
		localApiResult.Code = optcommon.ResultCode_PARMA_ERROR
		localApiResult.Msg = fmt.Sprintf("request addr is null")
		return localApiResult
	}
	reqUrl := fmt.Sprintf("%s/protocol/ac/%v", optAddr, apiId.Number())

	apiResultDto, err := utils.DoPbHttpRequest[*optcommon.APIResult](reqUrl, reqData)
	if nil == err {
		//响应正常
		if nil == apiResultDto.Data {
			localApiResult.Code = apiResultDto.Code
			localApiResult.Msg = apiResultDto.Msg
		} else {
			dist := reflect.New(reflect.TypeOf(t).Elem()).Interface().(proto.Message)
			err = anypb.UnmarshalTo(apiResultDto.Data, dist, proto.UnmarshalOptions{})
			if nil == err {
				localApiResult.Data = dist.(T)
				localApiResult.Code = apiResultDto.Code
				localApiResult.Msg = apiResultDto.Msg
			} else {
				//解析异常
				localApiResult.Msg = fmt.Sprintf("DoOperationRequest | Parsing exceptions, apiId=%v, err=%s", apiId.Number(), err.Error())
				localApiResult.Code = optcommon.ResultCode_INNER_ERROR
			}
		}
	} else {
		//响应异常
		localApiResult.Msg = fmt.Sprintf("DoOperationRequest | Call exceptions, apiId=%v, err=%s", apiId.Number(), err.Error())
		localApiResult.Code = optcommon.ResultCode_INNER_ERROR
	}
	if localApiResult.Code == optcommon.ResultCode_SUCCESS {
		localApiResult.IsSuccess = true
	} else {
		logger.Log.Error("DoOperationRequest | fail, apiId = %v, code = %v, msg = %s", apiId.Number(), localApiResult.Code.Number(), localApiResult.Msg)
	}
	return localApiResult
}

// DoOperationRequestAsync 异步调用运营平台接口请求
func DoOperationRequestAsync[T proto.Message](optAddr string, apiId optcommon.APIId, reqData proto.Message, callback func(*APIResult[T])) {
	taskName := fmt.Sprintf("DoOperationRequestAsync_%v", apiId.Number())
	utils.WorkerPool.Do(0, taskName, func() {
		apiResult := DoOperationRequest[T](optAddr, apiId, reqData)
		utils.WorkerPool.AddCallback(taskName, func() {
			if !apiResult.IsSuccess {
				logger.Log.Error("DoOperationRequestAsync | taskName[%s] fail, err[%s]", taskName, apiResult.Msg)
			}
			if nil != callback {
				callback(apiResult)
			}
		})
	})
}
